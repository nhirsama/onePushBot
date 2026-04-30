package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	base "github.com/nhirsama/onePushBot/internal/platform"
	model "github.com/nhirsama/onePushBot/internal/platform/telegram"
)

var (
	errBotStarted = errors.New("telegram bot client already started")
	errBotClosed  = errors.New("telegram bot client already closed")
)

// Client 定义 Telegram Bot 能力。
type Client interface {
	base.Client
	Native() *tgbotapi.BotAPI
	SendText(ctx context.Context, chatID string, text string) error
	ReplyText(ctx context.Context, chatID string, messageID string, text string) error
	SetCommands(ctx context.Context, commands []model.Command) error
}

// Dependencies 描述构造 Bot 客户端所需依赖。
type Dependencies struct {
	Logger base.Logger
	Native *tgbotapi.BotAPI
}

type client struct {
	cfg Config
	log base.Logger

	native *tgbotapi.BotAPI

	statusMu sync.RWMutex
	status   base.Status

	startMu sync.Mutex
	started bool
	closed  bool
	cancel  context.CancelFunc
	runWG   sync.WaitGroup

	accepting atomic.Bool

	eventCh chan base.Event
}

// New 创建 Telegram Bot 客户端。
func New(cfg Config, deps Dependencies) (Client, error) {
	cfg = cfg.withDefaults()
	if strings.TrimSpace(cfg.Token) == "" && deps.Native == nil {
		return nil, fmt.Errorf("telegram bot token 不能为空")
	}

	logger := deps.Logger
	if logger == nil {
		logger = base.NewDiscardLogger()
	}

	native := deps.Native
	if native == nil {
		bot, err := tgbotapi.NewBotAPI(cfg.Token)
		if err != nil {
			return nil, err
		}
		native = bot
	}

	return &client{
		cfg:     cfg,
		log:     logger,
		native:  native,
		status:  base.StatusStopped,
		eventCh: make(chan base.Event, cfg.EventBuffer),
	}, nil
}

func (c *client) Platform() base.Platform {
	return base.PlatformTelegramBot
}

func (c *client) Native() *tgbotapi.BotAPI {
	return c.native
}

func (c *client) Events() <-chan base.Event {
	return c.eventCh
}

func (c *client) Status() base.Status {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()
	return c.status
}

func (c *client) Start(ctx context.Context) error {
	c.startMu.Lock()
	defer c.startMu.Unlock()

	if c.closed {
		return errBotClosed
	}
	if c.started {
		return errBotStarted
	}

	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.started = true
	c.accepting.Store(true)
	c.setStatus(base.StatusStarting)

	c.runWG.Add(1)
	go c.run(runCtx)
	return nil
}

func (c *client) Close(ctx context.Context) error {
	c.startMu.Lock()
	if c.closed {
		c.startMu.Unlock()
		return nil
	}
	started := c.started
	c.started = false
	c.closed = true
	c.accepting.Store(false)
	cancel := c.cancel
	c.cancel = nil
	c.startMu.Unlock()

	if !started {
		c.setStatus(base.StatusClosed)
		return nil
	}

	if cancel != nil {
		cancel()
	}
	if c.native != nil {
		c.native.StopReceivingUpdates()
	}
	c.setStatus(base.StatusClosing)

	done := make(chan struct{})
	go func() {
		c.runWG.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	c.setStatus(base.StatusClosed)
	return nil
}

func (c *client) SendText(ctx context.Context, chatID string, text string) error {
	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(id, text)
	_, err = c.native.Send(msg)
	return err
}

func (c *client) ReplyText(ctx context.Context, chatID string, messageID string, text string) error {
	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return err
	}
	replyID, err := strconv.Atoi(messageID)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(id, text)
	msg.ReplyToMessageID = replyID
	_, err = c.native.Send(msg)
	return err
}

func (c *client) SetCommands(ctx context.Context, commands []model.Command) error {
	items := make([]tgbotapi.BotCommand, 0, len(commands))
	for _, item := range commands {
		items = append(items, tgbotapi.BotCommand{
			Command:     item.Command,
			Description: item.Description,
		})
	}
	_, err := c.native.Request(tgbotapi.NewSetMyCommands(items...))
	return err
}

func (c *client) run(ctx context.Context) {
	defer c.runWG.Done()
	defer c.accepting.Store(false)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = c.cfg.UpdateTimeout
	updateConfig.AllowedUpdates = c.cfg.AllowedUpdates

	updates := c.native.GetUpdatesChan(updateConfig)
	c.setStatus(base.StatusRunning)

	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			c.publishUpdate(update)
		}
	}
}

func (c *client) publishUpdate(update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		msg := update.Message
		subType := "message"
		if msg.IsCommand() {
			subType = "command"
		} else if msg.Chat.IsPrivate() {
			subType = "private_message"
		} else {
			subType = "group_message"
		}

		event := base.Event{
			ID:       strconv.Itoa(msg.MessageID),
			Platform: base.PlatformTelegramBot,
			Kind:     base.EventKindMessage,
			SubType:  subType,
			Time:     msg.Time(),
			Message: &base.Message{
				ID: strconv.Itoa(msg.MessageID),
				Chat: base.Chat{
					ID:   strconv.FormatInt(msg.Chat.ID, 10),
					Type: chatTypeFromChat(msg.Chat),
					Name: msg.Chat.Title,
				},
				Sender:       telegramUser(msg.From),
				Text:         msg.Text,
				Time:         msg.Time(),
				PlatformData: *msg,
			},
			Raw: update,
		}
		c.nonBlockingPublish(event)
	case update.CallbackQuery != nil:
		query := update.CallbackQuery
		chat := base.Chat{Type: base.ChatTypeUnknown}
		eventTime := time.Now()
		if query.Message != nil {
			chat = base.Chat{
				ID:   strconv.FormatInt(query.Message.Chat.ID, 10),
				Type: chatTypeFromChat(query.Message.Chat),
				Name: query.Message.Chat.Title,
			}
			eventTime = query.Message.Time()
		}

		event := base.Event{
			ID:       query.ID,
			Platform: base.PlatformTelegramBot,
			Kind:     base.EventKindNotice,
			SubType:  "callback_query",
			Time:     eventTime,
			Notice: &base.Notice{
				Type:         "callback_query",
				Chat:         chat,
				User:         telegramUser(query.From),
				PlatformData: *query,
			},
			Raw: update,
		}
		c.nonBlockingPublish(event)
	}
}

func (c *client) nonBlockingPublish(event base.Event) {
	if !c.accepting.Load() {
		return
	}

	select {
	case c.eventCh <- event:
	default:
	}
}

func (c *client) setStatus(status base.Status) {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()
	c.status = status
}

func chatTypeFromChat(chat *tgbotapi.Chat) base.ChatType {
	if chat == nil {
		return base.ChatTypeUnknown
	}
	switch {
	case chat.IsPrivate():
		return base.ChatTypePrivate
	case chat.IsChannel():
		return base.ChatTypeChannel
	default:
		return base.ChatTypeGroup
	}
}

func telegramUser(user *tgbotapi.User) base.User {
	if user == nil {
		return base.User{}
	}

	return base.User{
		ID:   strconv.FormatInt(user.ID, 10),
		Name: strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " ")),
	}
}

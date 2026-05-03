package feishu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
	larkdispatcher "github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	base "github.com/nhirsama/onePushBot/internal/platform"
)

var (
	errFeishuStarted = errors.New("feishu client already started")
	errFeishuClosed  = errors.New("feishu client already closed")
)

// Client 定义飞书平台能力。
type Client interface {
	base.Client
	Native() *lark.Client
	Handler() http.Handler
	SendText(ctx context.Context, receiveID string, text string) error
	SendCard(ctx context.Context, receiveID string, card any) error
}

// Dependencies 描述构造飞书客户端所需依赖。
type Dependencies struct {
	Logger base.Logger
	Native *lark.Client
}

type client struct {
	cfg Config
	log base.Logger

	native     *lark.Client
	dispatcher *larkdispatcher.EventDispatcher
	handler    http.Handler

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

type feishuTextContent struct {
	Text string `json:"text"`
}

// New 创建飞书平台客户端。
func New(cfg Config, deps Dependencies) (Client, error) {
	cfg = cfg.withDefaults()
	if strings.TrimSpace(cfg.AppID) == "" {
		return nil, fmt.Errorf("feishu app_id 不能为空")
	}
	if strings.TrimSpace(cfg.AppSecret) == "" {
		return nil, fmt.Errorf("feishu app_secret 不能为空")
	}

	logger := deps.Logger
	if logger == nil {
		logger = base.NewDiscardLogger()
	}

	native := deps.Native
	if native == nil {
		native = lark.NewClient(
			cfg.AppID,
			cfg.AppSecret,
			lark.WithReqTimeout(cfg.RequestTimeout),
		)
	}

	dispatcher := larkdispatcher.NewEventDispatcher(cfg.VerificationToken, cfg.EncryptKey)

	c := &client{
		cfg:        cfg,
		log:        logger,
		native:     native,
		dispatcher: dispatcher,
		status:     base.StatusStopped,
		eventCh:    make(chan base.Event, cfg.EventBuffer),
	}

	dispatcher.OnP2MessageReceiveV1(c.handleMessageReceive)
	dispatcher.OnP2MessageRecalledV1(c.handleMessageRecalled)
	dispatcher.OnP2MessageReadV1(c.handleMessageRead)
	c.handler = http.HandlerFunc(httpserverext.NewEventHandlerFunc(dispatcher))
	return c, nil
}

func (c *client) Platform() base.Platform {
	return base.PlatformFeishu
}

func (c *client) Native() *lark.Client {
	return c.native
}

func (c *client) Handler() http.Handler {
	return c.handler
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
		return errFeishuClosed
	}
	if c.started {
		return errFeishuStarted
	}

	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.started = true
	c.accepting.Store(true)
	c.setStatus(base.StatusRunning)

	c.runWG.Add(1)
	go func() {
		defer c.runWG.Done()
		<-runCtx.Done()
		c.accepting.Store(false)
	}()
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

func (c *client) SendText(ctx context.Context, receiveID string, text string) error {
	content, err := json.Marshal(feishuTextContent{Text: text})
	if err != nil {
		return err
	}

	body := larkim.NewCreateMessageReqBodyBuilder().
		ReceiveId(receiveID).
		MsgType("text").
		Content(string(content)).
		Build()
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(c.cfg.ReceiveIDType).
		Body(body).
		Build()

	resp, err := c.native.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("feishu send text failed: code=%d msg=%s", resp.Code, resp.Msg)
	}
	return nil
}

func (c *client) SendCard(ctx context.Context, receiveID string, card any) error {
	content, err := encodeCardContent(card)
	if err != nil {
		return err
	}

	body := larkim.NewCreateMessageReqBodyBuilder().
		ReceiveId(receiveID).
		MsgType("interactive").
		Content(content).
		Build()
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(c.cfg.ReceiveIDType).
		Body(body).
		Build()

	resp, err := c.native.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("feishu send card failed: code=%d msg=%s", resp.Code, resp.Msg)
	}
	return nil
}

func (c *client) handleMessageReceive(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	if event == nil || event.Event == nil || event.Event.Message == nil {
		return nil
	}

	msg := event.Event.Message
	sender := base.User{ID: extractUserID(event.Event.Sender)}
	if sender.ID != "" {
		sender.Name = sender.ID
	}

	text := extractMessageText(msg.MessageType, msg.Content)
	chatID := derefString(msg.ChatId)
	chatType := feishuChatType(msg.ChatType)
	timestamp := parseMilliTime(msg.CreateTime)
	messageID := derefString(msg.MessageId)
	subType := derefString(msg.MessageType)
	if subType == "" {
		subType = "message"
	}

	platformMessage := &base.Message{
		ID: messageID,
		Chat: base.Chat{
			ID:   chatID,
			Type: chatType,
		},
		Sender:       sender,
		Text:         text,
		RawText:      text,
		Time:         timestamp,
		DetailType:   subType,
		PlatformData: event,
	}

	c.publish(base.Event{
		ID:       messageID,
		Platform: base.PlatformFeishu,
		Kind:     base.EventKindMessage,
		SubType:  subType,
		Time:     timestamp,
		Message:  platformMessage,
		Raw:      event,
	})
	return nil
}

func (c *client) handleMessageRecalled(ctx context.Context, event *larkim.P2MessageRecalledV1) error {
	if event == nil || event.Event == nil {
		return nil
	}
	c.publish(base.Event{
		ID:       derefString(event.Event.MessageId),
		Platform: base.PlatformFeishu,
		Kind:     base.EventKindNotice,
		SubType:  "message_recalled",
		Time:     parseMilliTime(event.Event.RecallTime),
		Notice: &base.Notice{
			Type: "message_recalled",
			Chat: base.Chat{
				ID:   derefString(event.Event.ChatId),
				Type: base.ChatTypeGroup,
			},
			MessageID:    derefString(event.Event.MessageId),
			DetailType:   "message_recalled",
			PlatformData: event,
		},
		Raw: event,
	})
	return nil
}

func (c *client) handleMessageRead(ctx context.Context, event *larkim.P2MessageReadV1) error {
	if event == nil || event.Event == nil {
		return nil
	}
	messageID := ""
	if len(event.Event.MessageIdList) > 0 {
		messageID = event.Event.MessageIdList[0]
	}

	c.publish(base.Event{
		ID:       messageID,
		Platform: base.PlatformFeishu,
		Kind:     base.EventKindNotice,
		SubType:  "message_read",
		Time:     time.Now(),
		Notice: &base.Notice{
			Type:         "message_read",
			User:         base.User{ID: extractReaderID(event.Event.Reader)},
			MessageID:    messageID,
			DetailType:   "message_read",
			PlatformData: event,
		},
		Raw: event,
	})
	return nil
}

func (c *client) publish(event base.Event) {
	// Webhook 关闭后仍可能晚到，直接丢弃，避免继续向停用客户端投递事件。
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

func encodeCardContent(card any) (string, error) {
	switch value := card.(type) {
	case string:
		return value, nil
	case []byte:
		return string(value), nil
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(payload), nil
	}
}

func extractMessageText(messageType *string, content *string) string {
	if content == nil {
		return ""
	}
	if derefString(messageType) != "text" {
		return *content
	}

	var payload feishuTextContent
	if err := json.Unmarshal([]byte(*content), &payload); err != nil {
		return *content
	}
	return payload.Text
}

func extractUserID(sender *larkim.EventSender) string {
	if sender == nil || sender.SenderId == nil {
		return ""
	}
	if sender.SenderId.OpenId != nil && *sender.SenderId.OpenId != "" {
		return *sender.SenderId.OpenId
	}
	if sender.SenderId.UserId != nil && *sender.SenderId.UserId != "" {
		return *sender.SenderId.UserId
	}
	if sender.SenderId.UnionId != nil {
		return *sender.SenderId.UnionId
	}
	return ""
}

func extractReaderID(reader *larkim.EventMessageReader) string {
	if reader == nil {
		return ""
	}
	if reader.ReaderId == nil {
		return ""
	}
	if reader.ReaderId.OpenId != nil && *reader.ReaderId.OpenId != "" {
		return *reader.ReaderId.OpenId
	}
	if reader.ReaderId.UserId != nil && *reader.ReaderId.UserId != "" {
		return *reader.ReaderId.UserId
	}
	if reader.ReaderId.UnionId != nil {
		return *reader.ReaderId.UnionId
	}
	return ""
}

func feishuChatType(chatType *string) base.ChatType {
	switch derefString(chatType) {
	case "p2p":
		return base.ChatTypePrivate
	case "group", "topic_group":
		return base.ChatTypeGroup
	default:
		return base.ChatTypeUnknown
	}
}

func parseMilliTime(value *string) time.Time {
	if value == nil || *value == "" {
		return time.Now()
	}
	ms, err := strconv.ParseInt(*value, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.UnixMilli(ms)
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

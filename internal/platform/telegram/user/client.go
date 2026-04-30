package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotd/td/session"
	gotdtelegram "github.com/gotd/td/telegram"
	tgauth "github.com/gotd/td/telegram/auth"
	tgpeer "github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	base "github.com/nhirsama/onePushBot/internal/platform"
	model "github.com/nhirsama/onePushBot/internal/platform/telegram"
)

var errUserStarted = errors.New("telegram user client already started")

// MessageFilter 用于订阅用户态消息。
type MessageFilter struct {
	ChatIDs []string
}

// AuthProvider 为首次授权提供凭据。
type AuthProvider interface {
	Phone(ctx context.Context) (string, error)
	Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error)
	Password(ctx context.Context) (string, error)
}

// Client 定义 Telegram 用户态能力。
type Client interface {
	base.Client
	Native() *gotdtelegram.Client
	ListDialogs(ctx context.Context) ([]model.Dialog, error)
	GetHistory(ctx context.Context, chatID string, limit int) ([]model.Message, error)
	SubscribeMessages(ctx context.Context, filter MessageFilter) (<-chan model.Message, error)
}

// Dependencies 描述构造用户态客户端需要的依赖。
type Dependencies struct {
	Logger         base.Logger
	AuthProvider   AuthProvider
	SessionStorage session.Storage
}

type client struct {
	cfg  Config
	log  base.Logger
	auth AuthProvider

	statusMu sync.RWMutex
	status   base.Status

	native  *gotdtelegram.Client
	storage session.Storage

	eventCh chan base.Event

	startMu sync.Mutex
	started bool
	cancel  context.CancelFunc
	runWG   sync.WaitGroup

	readyOnce sync.Once
	readyCh   chan struct{}
	readyErr  error
	readyMu   sync.RWMutex

	cacheMu   sync.RWMutex
	users     map[int64]*tg.User
	chats     map[int64]*tg.Chat
	channels  map[int64]*tg.Channel
	peerCache map[string]tg.InputPeerClass
	sources   map[string]struct{}

	subMu      sync.RWMutex
	subs       map[uint64]*messageSubscription
	subCounter atomic.Uint64
}

type messageSubscription struct {
	filter MessageFilter
	ch     chan model.Message
}

type authenticator struct {
	provider AuthProvider
}

// New 创建 Telegram 用户态客户端。
func New(cfg Config, deps Dependencies) (Client, error) {
	cfg = cfg.withDefaults()
	if cfg.APIID <= 0 {
		return nil, fmt.Errorf("telegram api_id 必须大于 0")
	}
	if strings.TrimSpace(cfg.APIHash) == "" {
		return nil, fmt.Errorf("telegram api_hash 不能为空")
	}

	logger := deps.Logger
	if logger == nil {
		logger = base.NewDiscardLogger()
	}

	storage := deps.SessionStorage
	if storage == nil {
		if strings.TrimSpace(cfg.SessionPath) != "" {
			storage = &session.FileStorage{Path: cfg.SessionPath}
		} else {
			storage = &session.StorageMemory{}
		}
	}

	sources := make(map[string]struct{}, len(cfg.SourceChats))
	for _, id := range cfg.SourceChats {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			sources[trimmed] = struct{}{}
		}
	}

	c := &client{
		cfg:       cfg,
		log:       logger,
		auth:      deps.AuthProvider,
		status:    base.StatusStopped,
		storage:   storage,
		eventCh:   make(chan base.Event, cfg.EventBuffer),
		readyCh:   make(chan struct{}),
		users:     make(map[int64]*tg.User),
		chats:     make(map[int64]*tg.Chat),
		channels:  make(map[int64]*tg.Channel),
		peerCache: make(map[string]tg.InputPeerClass),
		sources:   sources,
		subs:      make(map[uint64]*messageSubscription),
	}

	dispatcher := tg.NewUpdateDispatcher()
	dispatcher.OnNewMessage(c.handleUpdateNewMessage)
	dispatcher.OnNewChannelMessage(c.handleUpdateNewChannelMessage)
	c.native = gotdtelegram.NewClient(cfg.APIID, cfg.APIHash, gotdtelegram.Options{
		SessionStorage: storage,
		UpdateHandler:  dispatcher,
	})
	return c, nil
}

func (c *client) Platform() base.Platform {
	return base.PlatformTelegramUser
}

func (c *client) Native() *gotdtelegram.Client {
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

	if c.started {
		return errUserStarted
	}
	if c.auth == nil {
		return fmt.Errorf("telegram 用户态客户端缺少授权提供器")
	}

	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.started = true
	c.setStatus(base.StatusStarting)

	c.runWG.Add(1)
	go c.run(runCtx)
	return nil
}

func (c *client) Close(ctx context.Context) error {
	c.startMu.Lock()
	if !c.started {
		c.startMu.Unlock()
		return nil
	}
	cancel := c.cancel
	c.cancel = nil
	c.started = false
	c.startMu.Unlock()

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

	c.closeSubscribers()
	close(c.eventCh)
	c.setStatus(base.StatusClosed)
	return nil
}

func (c *client) ListDialogs(ctx context.Context) ([]model.Dialog, error) {
	if err := c.waitReady(ctx); err != nil {
		return nil, err
	}

	items, err := dialogs.NewQueryBuilder(c.native.API()).
		GetDialogs().
		BatchSize(100).
		Collect(ctx)
	if err != nil {
		return nil, err
	}

	dialogsList := make([]model.Dialog, 0, len(items))
	for _, item := range items {
		c.rememberDialog(item)
		dialogsList = append(dialogsList, c.dialogFromElem(item))
	}
	return dialogsList, nil
}

func (c *client) GetHistory(ctx context.Context, chatID string, limit int) ([]model.Message, error) {
	if err := c.waitReady(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}

	peer, ok := c.findInputPeer(chatID)
	if !ok {
		if _, err := c.ListDialogs(ctx); err != nil {
			return nil, err
		}
		peer, ok = c.findInputPeer(chatID)
		if !ok {
			return nil, fmt.Errorf("未找到 chat_id=%s 对应的 Telegram peer", chatID)
		}
	}

	items, err := messages.NewQueryBuilder(c.native.API()).
		GetHistory(peer).
		BatchSize(min(limit, 100)).
		Collect(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Message, 0, len(items))
	for _, item := range items {
		c.rememberHistoryEntities(item.Entities, item.Peer)
		result = append(result, c.messageFromTG(item.Msg))
	}
	return result, nil
}

func (c *client) SubscribeMessages(ctx context.Context, filter MessageFilter) (<-chan model.Message, error) {
	if err := c.waitReady(ctx); err != nil {
		return nil, err
	}

	id := c.subCounter.Add(1)
	sub := &messageSubscription{
		filter: filter,
		ch:     make(chan model.Message, 64),
	}

	c.subMu.Lock()
	c.subs[id] = sub
	c.subMu.Unlock()

	go func() {
		<-ctx.Done()
		c.subMu.Lock()
		if current, ok := c.subs[id]; ok {
			delete(c.subs, id)
			close(current.ch)
		}
		c.subMu.Unlock()
	}()

	return sub.ch, nil
}

func (c *client) run(ctx context.Context) {
	defer c.runWG.Done()

	err := c.native.Run(ctx, func(runCtx context.Context) error {
		if err := c.ensureAuthorized(runCtx); err != nil {
			c.setReady(err)
			c.setStatus(base.StatusError)
			return err
		}

		if _, err := c.ListDialogs(runCtx); err != nil {
			c.log.Warn("Telegram 预热 dialogs 失败", "err", err)
		}

		c.setStatus(base.StatusRunning)
		c.setReady(nil)
		<-runCtx.Done()
		return runCtx.Err()
	})

	if err != nil && !errors.Is(err, context.Canceled) {
		c.log.Error("Telegram 用户态客户端退出", "err", err)
		c.setStatus(base.StatusError)
		c.setReady(err)
	}
}

func (c *client) ensureAuthorized(ctx context.Context) error {
	status, err := c.native.Auth().Status(ctx)
	if err != nil {
		return err
	}
	if status.Authorized {
		return nil
	}

	flow := tgauth.NewFlow(authenticator{provider: c.auth}, tgauth.SendCodeOptions{})
	return flow.Run(ctx, c.native.Auth())
}

func (c *client) handleUpdateNewMessage(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
	msg, ok := update.Message.AsNotEmpty()
	if !ok {
		return nil
	}
	return c.publishMessage(ctx, tgpeer.EntitiesFromUpdate(entities), msg, "message")
}

func (c *client) handleUpdateNewChannelMessage(ctx context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
	msg, ok := update.Message.AsNotEmpty()
	if !ok {
		return nil
	}
	return c.publishMessage(ctx, tgpeer.EntitiesFromUpdate(entities), msg, "channel_post")
}

func (c *client) publishMessage(ctx context.Context, entities tgpeer.Entities, msg tg.NotEmptyMessage, fallbackSubType string) error {
	c.rememberPeerEntities(entities)
	if peer, err := entities.ExtractPeer(msg.GetPeerID()); err == nil {
		if chatID, ok := peerClassID(msg.GetPeerID()); ok {
			c.cacheMu.Lock()
			c.peerCache[chatID] = peer
			c.cacheMu.Unlock()
		}
	}

	item := c.messageFromTG(msg)
	if !c.isAllowedSource(string(item.ChatID)) {
		return nil
	}

	subType := fallbackSubType
	switch item.Chat.Type {
	case base.ChatTypePrivate:
		subType = "private_message"
	case base.ChatTypeGroup:
		subType = "group_message"
	case base.ChatTypeChannel:
		subType = "channel_post"
	}

	event := base.Event{
		ID:       string(item.ID),
		Platform: base.PlatformTelegramUser,
		Kind:     base.EventKindMessage,
		SubType:  subType,
		Time:     item.Time,
		Message: &base.Message{
			ID:           string(item.ID),
			Chat:         item.Chat,
			Sender:       item.Sender,
			Text:         item.Text,
			Time:         item.Time,
			PlatformData: item.Raw,
		},
		Raw: item.Raw,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.eventCh <- event:
	}

	c.publishToSubscribers(item)
	return nil
}

func (c *client) publishToSubscribers(item model.Message) {
	c.subMu.RLock()
	defer c.subMu.RUnlock()

	for _, sub := range c.subs {
		if !matchMessageFilter(sub.filter, item) {
			continue
		}
		select {
		case sub.ch <- item:
		default:
		}
	}
}

func (c *client) dialogFromElem(item dialogs.Elem) model.Dialog {
	chatID, chatType, title := c.chatFromPeer(item.Last.GetPeerID())
	return model.Dialog{
		ID:    model.ChatID(chatID),
		Type:  chatType,
		Title: title,
		Raw:   item,
	}
}

func (c *client) messageFromTG(msg tg.NotEmptyMessage) model.Message {
	chatID, chatType, chatName := c.chatFromPeer(msg.GetPeerID())
	sender := c.senderFromMessage(msg)
	return model.Message{
		ID:     model.MessageID(strconv.Itoa(msg.GetID())),
		ChatID: model.ChatID(chatID),
		Chat: base.Chat{
			ID:   chatID,
			Type: chatType,
			Name: chatName,
		},
		Sender: sender,
		Text:   messageText(msg),
		Time:   time.Unix(int64(msg.GetDate()), 0),
		Raw:    msg,
	}
}

func (c *client) senderFromMessage(msg tg.NotEmptyMessage) base.User {
	from, ok := msg.GetFromID()
	if !ok {
		if author, ok := messagePostAuthor(msg); ok {
			chatID, _, _ := c.chatFromPeer(msg.GetPeerID())
			return base.User{ID: chatID, Name: author}
		}
		return base.User{}
	}

	switch peer := from.(type) {
	case *tg.PeerUser:
		c.cacheMu.RLock()
		user := c.users[peer.UserID]
		c.cacheMu.RUnlock()
		if user == nil {
			return base.User{ID: strconv.FormatInt(peer.UserID, 10)}
		}
		return base.User{
			ID:   strconv.FormatInt(user.ID, 10),
			Name: strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " ")),
		}
	case *tg.PeerChat:
		return base.User{ID: strconv.FormatInt(peer.ChatID, 10)}
	case *tg.PeerChannel:
		c.cacheMu.RLock()
		channel := c.channels[peer.ChannelID]
		c.cacheMu.RUnlock()
		if channel == nil {
			return base.User{ID: strconv.FormatInt(peer.ChannelID, 10)}
		}
		return base.User{ID: strconv.FormatInt(channel.ID, 10), Name: channel.Title}
	default:
		return base.User{}
	}
}

func (c *client) chatFromPeer(peer tg.PeerClass) (string, base.ChatType, string) {
	switch item := peer.(type) {
	case *tg.PeerUser:
		c.cacheMu.RLock()
		user := c.users[item.UserID]
		c.cacheMu.RUnlock()
		if user == nil {
			return strconv.FormatInt(item.UserID, 10), base.ChatTypePrivate, ""
		}
		return strconv.FormatInt(user.ID, 10), base.ChatTypePrivate, strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	case *tg.PeerChat:
		c.cacheMu.RLock()
		chat := c.chats[item.ChatID]
		c.cacheMu.RUnlock()
		if chat == nil {
			return strconv.FormatInt(item.ChatID, 10), base.ChatTypeGroup, ""
		}
		return strconv.FormatInt(chat.ID, 10), base.ChatTypeGroup, chat.Title
	case *tg.PeerChannel:
		c.cacheMu.RLock()
		channel := c.channels[item.ChannelID]
		c.cacheMu.RUnlock()
		if channel == nil {
			return strconv.FormatInt(item.ChannelID, 10), base.ChatTypeChannel, ""
		}
		return strconv.FormatInt(channel.ID, 10), base.ChatTypeChannel, channel.Title
	default:
		return "", base.ChatTypeUnknown, ""
	}
}

func (c *client) rememberDialog(item dialogs.Elem) {
	c.rememberPeerEntities(item.Entities)
	if chatID, ok := peerClassID(item.Last.GetPeerID()); ok {
		c.cacheMu.Lock()
		c.peerCache[chatID] = item.Peer
		c.cacheMu.Unlock()
	}
}

func (c *client) rememberHistoryEntities(entities tgpeer.Entities, peer tg.InputPeerClass) {
	c.rememberPeerEntities(entities)
	if id, ok := inputPeerID(peer); ok {
		c.cacheMu.Lock()
		c.peerCache[id] = peer
		c.cacheMu.Unlock()
	}
}

func (c *client) rememberPeerEntities(entities tgpeer.Entities) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	for id, user := range entities.Users() {
		c.users[id] = user
	}
	for id, chat := range entities.Chats() {
		c.chats[id] = chat
	}
	for id, channel := range entities.Channels() {
		c.channels[id] = channel
	}
}

func (c *client) findInputPeer(chatID string) (tg.InputPeerClass, bool) {
	c.cacheMu.RLock()
	if peer, ok := c.peerCache[chatID]; ok {
		c.cacheMu.RUnlock()
		return peer, true
	}

	id, err := strconv.ParseInt(chatID, 10, 64)
	if err == nil {
		if user, ok := c.users[id]; ok {
			c.cacheMu.RUnlock()
			return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, true
		}
		if chat, ok := c.chats[id]; ok {
			c.cacheMu.RUnlock()
			return &tg.InputPeerChat{ChatID: chat.ID}, true
		}
		if channel, ok := c.channels[id]; ok {
			c.cacheMu.RUnlock()
			return &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}, true
		}
	}
	c.cacheMu.RUnlock()
	return nil, false
}

func (c *client) isAllowedSource(chatID string) bool {
	if len(c.sources) == 0 {
		return true
	}
	_, ok := c.sources[chatID]
	return ok
}

func (c *client) waitReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.readyCh:
		c.readyMu.RLock()
		defer c.readyMu.RUnlock()
		return c.readyErr
	}
}

func (c *client) setReady(err error) {
	c.readyOnce.Do(func() {
		c.readyMu.Lock()
		c.readyErr = err
		c.readyMu.Unlock()
		close(c.readyCh)
	})
}

func (c *client) setStatus(status base.Status) {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()
	c.status = status
}

func (c *client) closeSubscribers() {
	c.subMu.Lock()
	defer c.subMu.Unlock()

	for id, sub := range c.subs {
		close(sub.ch)
		delete(c.subs, id)
	}
}

func matchMessageFilter(filter MessageFilter, item model.Message) bool {
	if len(filter.ChatIDs) == 0 {
		return true
	}
	for _, id := range filter.ChatIDs {
		if id == string(item.ChatID) {
			return true
		}
	}
	return false
}

func peerClassID(peer tg.PeerClass) (string, bool) {
	switch item := peer.(type) {
	case *tg.PeerUser:
		return strconv.FormatInt(item.UserID, 10), true
	case *tg.PeerChat:
		return strconv.FormatInt(item.ChatID, 10), true
	case *tg.PeerChannel:
		return strconv.FormatInt(item.ChannelID, 10), true
	default:
		return "", false
	}
}

func inputPeerID(peer tg.InputPeerClass) (string, bool) {
	switch item := peer.(type) {
	case *tg.InputPeerUser:
		return strconv.FormatInt(item.UserID, 10), true
	case *tg.InputPeerChat:
		return strconv.FormatInt(item.ChatID, 10), true
	case *tg.InputPeerChannel:
		return strconv.FormatInt(item.ChannelID, 10), true
	default:
		return "", false
	}
}

func (a authenticator) Phone(ctx context.Context) (string, error) {
	return a.provider.Phone(ctx)
}

func (a authenticator) Password(ctx context.Context) (string, error) {
	return a.provider.Password(ctx)
}

func (a authenticator) AcceptTermsOfService(context.Context, tg.HelpTermsOfService) error {
	return fmt.Errorf("当前未实现 Telegram 用户态自动注册流程")
}

func (a authenticator) SignUp(context.Context) (tgauth.UserInfo, error) {
	return tgauth.UserInfo{}, fmt.Errorf("当前未实现 Telegram 用户态自动注册流程")
}

func (a authenticator) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	return a.provider.Code(ctx, sentCode)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func messageText(msg tg.NotEmptyMessage) string {
	switch item := msg.(type) {
	case *tg.Message:
		return item.Message
	case *tg.MessageService:
		return ""
	default:
		return ""
	}
}

func messagePostAuthor(msg tg.NotEmptyMessage) (string, bool) {
	switch item := msg.(type) {
	case *tg.Message:
		return item.GetPostAuthor()
	default:
		return "", false
	}
}

package napcat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/qq"
)

var (
	errClientStarted  = errors.New("napcat client already started")
	errClientClosed   = errors.New("napcat client already closed")
	errClientNotReady = errors.New("napcat client is not connected")
)

// Dependencies 描述 NapCat 客户端构造所需的依赖。
type Dependencies struct {
	Logger base.Logger
	Dialer *websocket.Dialer
	Now    func() time.Time
}

type client struct {
	cfg    Config
	log    base.Logger
	dialer *websocket.Dialer
	now    func() time.Time

	statusMu sync.RWMutex
	status   base.Status

	eventCh chan base.Event
	doneCh  chan struct{}

	startMu sync.Mutex
	started bool
	closed  bool
	cancel  context.CancelFunc
	runWG   sync.WaitGroup

	connMu  sync.RWMutex
	writeMu sync.Mutex
	conn    *websocket.Conn

	echoCounter   atomic.Uint64
	lastHeartbeat atomic.Int64
	connected     atomic.Bool
	responseMap   sync.Map
}

// New 创建一个基于 NapCat WebSocket 的 QQ 客户端。
func New(cfg Config, deps Dependencies) (qq.Client, error) {
	cfg = cfg.withDefaults()
	if strings.TrimSpace(cfg.APIURL) == "" {
		return nil, fmt.Errorf("napcat api url 不能为空")
	}

	logger := deps.Logger
	if logger == nil {
		logger = base.NewDiscardLogger()
	}

	dialer := deps.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}

	now := deps.Now
	if now == nil {
		now = time.Now
	}

	return &client{
		cfg:     cfg,
		log:     logger,
		dialer:  dialer,
		now:     now,
		status:  base.StatusStopped,
		eventCh: make(chan base.Event, cfg.EventBuffer),
		doneCh:  make(chan struct{}),
	}, nil
}

func (c *client) Platform() base.Platform {
	return base.PlatformQQ
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
		return errClientClosed
	}
	if c.started {
		return errClientStarted
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
	if c.closed {
		c.startMu.Unlock()
		return nil
	}
	started := c.started
	c.started = false
	c.closed = true
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
	c.closeCurrentConn()

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

	c.finishClose()
	return nil
}

func (c *client) Call(ctx context.Context, action string, params any) (json.RawMessage, error) {
	if deadline, ok := ctx.Deadline(); !ok || deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.cfg.RequestTimeout)
		defer cancel()
	}

	echo := fmt.Sprintf("%d-%d", c.now().UnixNano(), c.echoCounter.Add(1))
	replyCh := make(chan rawEnvelope, 1)
	c.responseMap.Store(echo, replyCh)
	defer c.responseMap.Delete(echo)

	payload, err := json.Marshal(rawRequest{
		Action: action,
		Echo:   echo,
		Params: params,
	})
	if err != nil {
		return nil, fmt.Errorf("序列化 napcat 请求失败: %w", err)
	}

	if err := c.write(ctx, payload); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case reply := <-replyCh:
		if reply.Status != "" && reply.Status != "ok" {
			return nil, fmt.Errorf("napcat action %s 失败: retcode=%d", action, reply.RetCode)
		}
		return reply.Data, nil
	}
}

func (c *client) SendGroupText(ctx context.Context, groupID string, text string) error {
	_, err := c.Call(ctx, "send_group_msg", map[string]any{
		"group_id": toNapcatID(groupID),
		"message": []map[string]any{
			{
				"type": "text",
				"data": map[string]any{"text": text},
			},
		},
	})
	return err
}

func (c *client) SendPrivateText(ctx context.Context, userID string, text string) error {
	_, err := c.Call(ctx, "send_private_msg", map[string]any{
		"user_id": toNapcatID(userID),
		"message": []map[string]any{
			{
				"type": "text",
				"data": map[string]any{"text": text},
			},
		},
	})
	return err
}

func (c *client) SendLike(ctx context.Context, userID string, times int) error {
	_, err := c.Call(ctx, "send_like", map[string]any{
		"user_id": toNapcatID(userID),
		"times":   times,
	})
	return err
}

func (c *client) SendPoke(ctx context.Context, groupID string, userID string) error {
	params := map[string]any{
		"group_id":  toNapcatID(groupID),
		"target_id": toNapcatID(userID),
	}
	if c.cfg.SelfID != "" {
		params["user_id"] = toNapcatID(c.cfg.SelfID)
	}
	_, err := c.Call(ctx, "send_poke", params)
	return err
}

func (c *client) SetMsgEmojiLike(ctx context.Context, messageID string, emojiID int, set bool) error {
	_, err := c.Call(ctx, "set_msg_emoji_like", map[string]any{
		"message_id": messageID,
		"emoji_id":   emojiID,
		"set":        set,
	})
	return err
}

func (c *client) GetGroupMemberInfo(ctx context.Context, groupID string, userID string, noCache bool) (*base.GroupMemberInfo, error) {
	data, err := c.Call(ctx, "get_group_member_info", map[string]any{
		"group_id": toNapcatID(groupID),
		"user_id":  toNapcatID(userID),
		"no_cache": noCache,
	})
	if err != nil {
		return nil, err
	}

	var member rawGroupMemberInfo
	if err := json.Unmarshal(data, &member); err != nil {
		return nil, fmt.Errorf("解析群成员信息失败: %w", err)
	}
	return member.toDomain(), nil
}

func toNapcatID(id string) any {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return id
	}
	if value, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return value
	}
	return id
}

func (m rawGroupMemberInfo) toDomain() *base.GroupMemberInfo {
	return &base.GroupMemberInfo{
		GroupID:         normalizeID(m.GroupID),
		UserID:          normalizeID(m.UserID),
		Nickname:        m.Nickname,
		Card:            m.Card,
		Sex:             m.Sex,
		Age:             m.Age,
		JoinTime:        m.JoinTime,
		LastSentTime:    m.LastSentTime,
		Level:           m.Level,
		QQLevel:         m.QQLevel,
		Role:            m.Role,
		Title:           m.Title,
		Area:            m.Area,
		Unfriendly:      m.Unfriendly,
		TitleExpireTime: m.TitleExpireTime,
		CardChangeable:  m.CardChangeable,
		ShutUpTimestamp: m.ShutUpTimestamp,
		IsRobot:         m.IsRobot,
		QAge:            m.QAge,
	}
}

func (c *client) run(ctx context.Context) {
	defer c.runWG.Done()
	defer close(c.doneCh)

	if c.cfg.HeartbeatTimeout > 0 {
		c.runWG.Add(1)
		go c.monitorHeartbeat(ctx)
	}

	attempt := 0
	backoff := time.Second

	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := c.connect(ctx)
		if err != nil {
			attempt++
			c.setStatus(base.StatusError)
			c.log.Error("NapCat 连接失败", "attempt", attempt, "err", err)
			if c.cfg.ReconnectMaxAttempts > 0 && attempt >= c.cfg.ReconnectMaxAttempts {
				return
			}
			if !sleepWithContext(ctx, backoff) {
				return
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		attempt = 0
		backoff = time.Second
		c.lastHeartbeat.Store(c.now().UnixNano())
		c.setConn(conn)
		c.setStatus(base.StatusRunning)
		c.log.Info("NapCat 已连接")

		if err := c.readLoop(ctx, conn); err != nil && ctx.Err() == nil {
			c.log.Warn("NapCat 连接中断，准备重连", "err", err)
		}

		c.clearConn(conn)
		_ = conn.Close()
	}
}

func (c *client) monitorHeartbeat(ctx context.Context) {
	defer c.runWG.Done()

	interval := c.cfg.HeartbeatTimeout / 2
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !c.connected.Load() {
				continue
			}
			last := c.lastHeartbeat.Load()
			if last == 0 {
				continue
			}
			if c.now().Sub(time.Unix(0, last)) > c.cfg.HeartbeatTimeout {
				c.log.Warn("NapCat 心跳超时，主动断开等待重连")
				c.closeCurrentConn()
			}
		}
	}
}

func (c *client) connect(ctx context.Context) (*websocket.Conn, error) {
	targets, err := buildDialTargets(c.cfg.APIURL, c.cfg.Token)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, target := range targets {
		dialCtx, cancel := context.WithTimeout(ctx, c.cfg.DialTimeout)
		conn, _, err := c.dialer.DialContext(dialCtx, target, nil)
		cancel()
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("未找到可用的 napcat 连接地址")
	}
	return nil, lastErr
}

func (c *client) readLoop(ctx context.Context, conn *websocket.Conn) error {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var envelope rawEnvelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			c.log.Warn("解析 NapCat 消息失败", "err", err)
			continue
		}

		if c.dispatchResponse(envelope) {
			continue
		}
		if c.handleProtocolEvent(envelope) {
			continue
		}

		event, ok, err := mapEnvelopeToEvent(envelope, c.cfg.SelfID)
		if err != nil {
			c.log.Warn("映射 NapCat 事件失败", "err", err)
			continue
		}
		if !ok {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case c.eventCh <- event:
		}
	}
}

func (c *client) handleProtocolEvent(envelope rawEnvelope) bool {
	if envelope.PostType != "meta_event" {
		return false
	}
	if envelope.MetaEventType == "heartbeat" {
		c.lastHeartbeat.Store(c.now().UnixNano())
	}
	return true
}

func (c *client) dispatchResponse(envelope rawEnvelope) bool {
	if envelope.Echo == "" {
		return false
	}
	value, ok := c.responseMap.Load(envelope.Echo)
	if !ok {
		return false
	}

	replyCh, ok := value.(chan rawEnvelope)
	if !ok {
		return false
	}

	select {
	case replyCh <- envelope:
	default:
	}
	return true
}

func (c *client) write(ctx context.Context, payload []byte) error {
	conn := c.currentConn()
	if conn == nil {
		return errClientNotReady
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetWriteDeadline(deadline); err != nil {
			return fmt.Errorf("设置 napcat 写超时失败: %w", err)
		}
	} else if err := conn.SetWriteDeadline(c.now().Add(c.cfg.WriteTimeout)); err != nil {
		return fmt.Errorf("设置 napcat 写超时失败: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		c.closeCurrentConn()
		return fmt.Errorf("写入 napcat 失败: %w", err)
	}
	return nil
}

func (c *client) setConn(conn *websocket.Conn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	c.conn = conn
	c.connected.Store(true)
}

func (c *client) clearConn(conn *websocket.Conn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn == conn {
		c.conn = nil
		c.connected.Store(false)
	}
}

func (c *client) closeCurrentConn() {
	c.connMu.Lock()
	conn := c.conn
	c.conn = nil
	c.connected.Store(false)
	c.connMu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}

func (c *client) currentConn() *websocket.Conn {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.conn
}

func (c *client) setStatus(status base.Status) {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()
	c.status = status
}

func (c *client) finishClose() {
	c.setStatus(base.StatusClosed)
}

func buildDialTargets(rawTarget, token string) ([]string, error) {
	target := strings.TrimSpace(rawTarget)
	if target == "" {
		return nil, fmt.Errorf("napcat api url 不能为空")
	}

	hasScheme := strings.Contains(target, "://")
	var (
		parsed *url.URL
		err    error
	)
	if hasScheme {
		parsed, err = url.Parse(target)
	} else {
		parsed, err = url.Parse("//" + target)
	}
	if err != nil {
		return nil, fmt.Errorf("解析 napcat url 失败: %w", err)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("napcat url 无效: %s", rawTarget)
	}

	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/ws"
	}

	query := parsed.Query()
	if token != "" && query.Get("access_token") == "" {
		query.Set("access_token", token)
	}
	parsed.RawQuery = query.Encode()

	schemes := []string{"wss", "ws"}
	if hasScheme {
		switch parsed.Scheme {
		case "ws":
			schemes = []string{"ws", "wss"}
		case "wss":
			schemes = []string{"wss", "ws"}
		case "http":
			schemes = []string{"ws", "wss"}
		case "https":
			schemes = []string{"wss", "ws"}
		default:
			return nil, fmt.Errorf("不支持的 napcat 协议: %s", parsed.Scheme)
		}
	}

	targets := make([]string, 0, len(schemes))
	seen := make(map[string]struct{}, len(schemes))
	for _, scheme := range schemes {
		item := *parsed
		item.Scheme = scheme
		full := item.String()
		if _, ok := seen[full]; ok {
			continue
		}
		seen[full] = struct{}{}
		targets = append(targets, full)
	}
	return targets, nil
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

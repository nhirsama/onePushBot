package napcat

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nhirsama/onePushBot/internal/domain"
	pkgDomain "github.com/nhirsama/onePushBot/pkg/domain"
)

// ConnectionManager 连接管理器
type ConnectionManager struct {
	*domain.BotConfig
	conn        *websocket.Conn
	connMu      sync.RWMutex
	connected   atomic.Bool
	connectTime time.Time
	log         pkgDomain.Log

	reconnectCount atomic.Int32
	// 心跳
	heartbeatChan chan struct{}
	heartbeatCb   func()
	// 控制
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	// 读写通道 & 启动标记
	readCh  chan []byte
	writeCh chan []byte
	ioOnce  sync.Once // 确保读写协程只启动一次
}

func (c *ConnectionManager) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.connected.Load() {
		return nil
	}

	// 尝试连接

	conn, err := c.login(ctx)
	if err != nil {
		return fmt.Errorf("连接 websocket: %w", err)
	}
	c.conn = conn
	c.connected.Store(true)
	c.connectTime = time.Now()

	c.log.Info("已连接到 WebSocket")

	// 确保读写协程只启动一次
	c.ioOnce.Do(func() {
		c.wg.Add(2)
		go c.readLoop()
		go c.writeLoop()
	})
	return nil
}

func (c *ConnectionManager) login(ctx context.Context) (*websocket.Conn, error) {
	host := c.BotConfig.ApiUrl
	if host == "" {
		return nil, fmt.Errorf("ApiUrl 不能为空")
	}

	token := c.BotConfig.Token

	q := url.Values{}
	if token != "" {
		q.Set("access_token", token)
	}
	rawQuery := q.Encode()

	dial := func(scheme string) (*websocket.Conn, error) {
		u := url.URL{
			Scheme:   scheme,
			Host:     host,
			Path:     "/ws",
			RawQuery: rawQuery,
		}
		fullURL := u.String()

		conn, _, err := websocket.DefaultDialer.DialContext(ctx, fullURL, nil)
		if err != nil {
			c.log.Error("WebSocket 连接失败", "url", fullURL, "err", err)
			return nil, err
		}
		c.log.Info("WebSocket 已连接", "url", fullURL)
		return conn, nil
	}

	var lastErr error
	// 无视用户给的协议，固定先试 wss，再试 ws
	for _, scheme := range []string{"wss", "ws"} {
		if conn, err := dial(scheme); err == nil {
			return conn, nil
		} else {
			lastErr = err
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("WebSocket 连接失败: %w", lastErr)
	}
	return nil, fmt.Errorf("WebSocket 连接失败")
}

func (c *ConnectionManager) Reconnect(ctx context.Context) error {
	attempts := int64(0)
	maxAttempts := c.ReconnectMaxAttempts
	interval := 1 * time.Second

	for {
		if maxAttempts > 0 && attempts >= maxAttempts {
			return fmt.Errorf("已达到最大重连尝试次数： %d", attempts)
		}

		attempts++
		c.reconnectCount.Add(1)

		c.log.Info(fmt.Sprintf("重新连接中...(重试 %d 次)", attempts))

		if err := c.Connect(ctx); err == nil {
			c.log.Info(fmt.Sprintf("在%d次尝试后重新连接成功", attempts))
			return nil
		} else {
			c.log.Debug("重新连接时失败", "err", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			// 指数退避
			if interval < 30*time.Second {
				interval *= 2
			}
		}
	}
}

func (c *ConnectionManager) Close() error {
	c.cancel()
	c.ctx.Done()
	c.wg.Wait()

	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		if err == nil {
			c.conn = nil
			c.connected.Store(false)
			return nil
		}
		c.connected.Store(false)
		return err
	}

	return nil
}

func (c *ConnectionManager) IsConnected() bool {
	return c.connected.Load()
}

func (c *ConnectionManager) GetConnectTime() time.Time {
	return c.connectTime
}

func (c *ConnectionManager) GetReconnectCount() int {
	return int(c.reconnectCount.Load())
}

func (c *ConnectionManager) StartHeartbeat(ctx context.Context) {
	go c.heartbeatLoop(ctx)
}

func (c *ConnectionManager) OnHeartbeat(callback func()) {
	c.heartbeatCb = callback
}

func (c *ConnectionManager) NotifyHeartbeat() {
	select {
	case c.heartbeatChan <- struct{}{}:
	default:
	}
}

// heartbeatLoop 心跳监控循环
func (c *ConnectionManager) heartbeatLoop(ctx context.Context) {
	c.wg.Add(1)
	timer := time.NewTimer(time.Duration(c.HeartbeatTimeout) * time.Second)
	defer func() {
		timer.Stop()
		c.wg.Done()
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.heartbeatChan:
			// 收到心跳，重置定时器
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(time.Duration(c.HeartbeatTimeout) * time.Second)

			if c.heartbeatCb != nil {
				func() {
					defer func() {
						if r := recover(); r != nil {
							c.log.Error("心跳回调发生 panic", "error", r)
						}
					}()
					c.heartbeatCb()
				}()
			}
		case <-timer.C:
			// 心跳超时，尝试重连
			c.log.Warn("心跳超时，正在重新连接")
			c.connected.Store(false)
			if err := c.Reconnect(c.ctx); err != nil {
				c.log.Error("重新连接失败", "err", err)
			}
			// 重新连接后重置定时器
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(time.Duration(c.HeartbeatTimeout) * time.Second)
		}
	}
}

func NewConnectionManager(parentCtx context.Context, log pkgDomain.Log, config *domain.BotConfig) domain.ConnectionManager {
	ctx, cancel := context.WithCancel(parentCtx)
	c := &ConnectionManager{
		BotConfig:     config,
		heartbeatChan: make(chan struct{}, 1),
		ctx:           ctx,
		cancel:        cancel,
		log:           log,
		readCh:        make(chan []byte, 128),
		writeCh:       make(chan []byte, 128),
	}
	return c
}

// ReadChan 返回只读通道，上层通过 <-ReadChan() 获取从 WebSocket 收到的字节流
func (c *ConnectionManager) ReadChan() <-chan []byte {
	return c.readCh
}

// WriteChan 返回写通道，上层通过 WriteChan() <- data 发送字节流到 WebSocket
func (c *ConnectionManager) WriteChan() chan<- []byte {
	return c.writeCh
}

// readLoop 持续从 WebSocket 连接读取消息，并写入 readCh。
// 只在获取 conn 指针时短暂加读锁，真正的 ReadMessage 不持锁，避免阻塞其他方法。
func (c *ConnectionManager) readLoop() {
	defer c.wg.Done()
	c.log.Debug("WebSocket 读协程已启动")

	timeout := time.Duration(c.HeartbeatTimeout) * time.Second

	for {
		// 全局 context 被取消时退出
		select {
		case <-c.ctx.Done():
			c.log.Info("读协程收到关闭信号，退出")
			return
		default:
		}

		// 取当前连接（短暂加锁）
		c.connMu.RLock()
		conn := c.conn
		c.connMu.RUnlock()

		if conn == nil || !c.connected.Load() {
			// 没有有效连接，稍等再试
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// 设置读超时，防止 ReadMessage 无限阻塞
		_ = conn.SetReadDeadline(time.Now().Add(timeout))

		_, msg, err := conn.ReadMessage()
		if err != nil {
			c.log.Warn("读协程读取 WebSocket 消息失败，等待重连", "err", err)
			// 这里不直接退出，让重连逻辑重新建立连接后继续读
			c.connected.Store(false)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// 将读取到的字节流发送到 readCh，支持上层用 select 做超时
		select {
		case c.readCh <- msg:
		case <-c.ctx.Done():
			return
		}
	}
}

// writeLoop 从 writeCh 读取待发送的字节流，并写入 WebSocket。
// 使用单写者模型，保证 gorilla/websocket 的并发安全要求。
func (c *ConnectionManager) writeLoop() {
	defer c.wg.Done()
	c.log.Info("WebSocket 写协程已启动")

	timeout := time.Duration(c.HeartbeatTimeout) * time.Second

	for {
		select {
		case <-c.ctx.Done():
			c.log.Debug("写协程收到关闭信号，退出")
			return

		case data := <-c.writeCh:
			// 取当前连接（短暂加锁）
			c.connMu.RLock()
			conn := c.conn
			c.connMu.RUnlock()

			if conn == nil || !c.connected.Load() {
				c.log.Warn("写协程检测到连接不可用，丢弃或等待重连")
				// 这里你可以选择缓存 / 重试，目前简单丢弃
				continue
			}

			_ = conn.SetWriteDeadline(time.Now().Add(timeout))

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				c.log.Error("写协程发送 WebSocket 消息失败", "err", err)
				c.connected.Store(false)
				// 等待外部重连后再继续
				continue
			}
		}
	}
}

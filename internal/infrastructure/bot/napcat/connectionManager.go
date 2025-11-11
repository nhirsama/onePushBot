package napcat

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nhirsama/onePushBot/internal/domain"
	pkgDomain "github.com/nhirsama/onePushBot/pkg/domain"
)

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
}

func (c *ConnectionManager) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.connected.Load() {
		return nil
	}

	// 尝试连接
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.BotConfig.ApiUrl, nil)
	if err != nil {
		return fmt.Errorf("连接 websocket: %w", err)
	}
	c.conn = conn
	c.connected.Store(true)
	c.connectTime = time.Now()

	c.log.Info("已连接到 WebSocket")
	c.log.Debug(fmt.Sprintf("已连接到:%s", c.ApiUrl))

	return nil
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
	go c.heartbeatLoop()
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
func (c *ConnectionManager) heartbeatLoop() {
	c.wg.Add(1)
	ticker := time.NewTimer(time.Duration(c.HeartbeatTimeout))
	defer func() {
		ticker.Stop()
		c.wg.Done()
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.heartbeatChan:
			// 收到心跳，重置定时器
			if !ticker.Stop() {
				select {
				case <-ticker.C:
				default:
				}
			}
			ticker.Reset(time.Duration(c.HeartbeatTimeout))

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
		case <-ticker.C:
			// 心跳超时，尝试重连
			c.log.Warn("心跳超时，正在重新连接")
			c.connected.Store(false)
			if err := c.Reconnect(c.ctx); err != nil {
				c.log.Error("重新连接失败", "err", err)
			}
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
	}
	//c.StartHeartbeat(ctx)
	return c
}

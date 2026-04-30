package platform

import (
	"context"
	"sync"
	"time"
)

// Hub 把平台客户端和统一总线连接起来。
type Hub interface {
	Register(client Client) error
	Start(ctx context.Context) error
	Close(ctx context.Context) error
	Bus() Bus
	Get(platform Platform) (Client, bool)
	All() []Client
}

type hub struct {
	bus      Bus
	registry Registry

	mu      sync.Mutex
	started bool
	closed  bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewHub 创建一个默认使用内存总线的平台 Hub。
func NewHub(bus Bus) Hub {
	if bus == nil {
		bus = NewMemoryBus()
	}
	return &hub{
		bus:      bus,
		registry: NewRegistry(),
	}
}

func (h *hub) Register(client Client) error {
	h.mu.Lock()
	started := h.started
	closed := h.closed
	h.mu.Unlock()
	if closed {
		return ErrHubClosed
	}
	if started {
		return ErrHubStarted
	}
	return h.registry.Register(client)
}

func (h *hub) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return ErrHubClosed
	}
	if h.started {
		h.mu.Unlock()
		return ErrHubStarted
	}
	runCtx, cancel := context.WithCancel(ctx)
	h.started = true
	h.cancel = cancel
	h.mu.Unlock()

	startedClients := make([]Client, 0)
	for _, client := range h.registry.All() {
		if err := client.Start(runCtx); err != nil {
			cancel()
			closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
			for _, startedClient := range startedClients {
				_ = startedClient.Close(closeCtx)
			}
			closeCancel()
			h.wg.Wait()
			h.mu.Lock()
			h.started = false
			h.cancel = nil
			h.mu.Unlock()
			return err
		}

		startedClients = append(startedClients, client)
		h.wg.Add(1)
		go h.forward(runCtx, client)
	}

	return nil
}

func (h *hub) Close(ctx context.Context) error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	h.started = false
	cancel := h.cancel
	h.cancel = nil
	h.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	for _, client := range h.registry.All() {
		if err := client.Close(ctx); err != nil {
			return err
		}
	}

	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	return h.bus.Close()
}

func (h *hub) Bus() Bus {
	return h.bus
}

func (h *hub) Get(platform Platform) (Client, bool) {
	return h.registry.Get(platform)
}

func (h *hub) All() []Client {
	return h.registry.All()
}

func (h *hub) forward(ctx context.Context, client Client) {
	defer h.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-client.Events():
			if !ok {
				return
			}
			_ = h.bus.Publish(ctx, event)
		}
	}
}

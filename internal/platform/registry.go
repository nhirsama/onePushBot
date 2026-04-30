package platform

import "sync"

// Registry 保存平台客户端实例，供上层统一注册和查询。
type Registry interface {
	Register(client Client) error
	Get(platform Platform) (Client, bool)
	All() []Client
}

type memoryRegistry struct {
	mu      sync.RWMutex
	clients map[Platform]Client
}

// NewRegistry 创建一个进程内平台注册表。
func NewRegistry() Registry {
	return &memoryRegistry{
		clients: make(map[Platform]Client),
	}
}

func (r *memoryRegistry) Register(client Client) error {
	if client == nil {
		return ErrClientNil
	}
	if client.Platform() == "" {
		return ErrClientInvalid
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[client.Platform()]; ok {
		return ErrClientExists
	}
	r.clients[client.Platform()] = client
	return nil
}

func (r *memoryRegistry) Get(platform Platform) (Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, ok := r.clients[platform]
	return client, ok
}

func (r *memoryRegistry) All() []Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Client, 0, len(r.clients))
	for _, client := range r.clients {
		result = append(result, client)
	}
	return result
}

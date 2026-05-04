package platformclient

import base "github.com/nhirsama/onePushBot/internal/platform"

// Source is the narrow client lookup surface pkg modules need from the runtime.
type Source interface {
	Get(platform base.Platform) (any, bool)
}

type enumerableSource interface {
	All() []any
}

type runtimeSource interface {
	Get(platform base.Platform) (base.Client, bool)
	All() []base.Client
}

type sourceView struct {
	runtime runtimeSource
}

func NewSource(runtime runtimeSource) Source {
	if runtime == nil {
		return nil
	}
	return sourceView{runtime: runtime}
}

func (s sourceView) Get(platform base.Platform) (any, bool) {
	return s.runtime.Get(platform)
}

func (s sourceView) All() []any {
	clients := s.runtime.All()
	result := make([]any, 0, len(clients))
	for _, client := range clients {
		result = append(result, client)
	}
	return result
}

func Get[T any](source Source, platform base.Platform) (T, bool) {
	var zero T
	if source == nil {
		return zero, false
	}
	client, ok := source.Get(platform)
	if !ok {
		return zero, false
	}
	capability, ok := client.(T)
	return capability, ok
}

func ForEvent[T any](source Source, event base.Event) (T, bool) {
	return Get[T](source, event.Platform)
}

func First[T any](source Source) (T, bool) {
	var zero T
	if source == nil {
		return zero, false
	}
	enumerable, ok := source.(enumerableSource)
	if !ok {
		return zero, false
	}
	for _, client := range enumerable.All() {
		capability, ok := client.(T)
		if ok {
			return capability, true
		}
	}
	return zero, false
}

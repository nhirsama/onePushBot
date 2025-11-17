package domain

import (
	"context"

	"github.com/asaskevich/EventBus"
)

type ApiDistribute interface {
	Call(context.Context, string, any) ([]byte, error)
	Start(*EventBus.Bus) error
}

package modules

import (
	"context"

	"github.com/nhirsama/onePushBot/pkg/nowcoder_tracker"
)

func init() {
	register(Module{
		Name: "nowcoder_daily",
		Start: func(ctx context.Context, deps Dependencies) error {
			nowcoderTracker.StartDaily(ctx, deps.Clients)
			return nil
		},
	})
}

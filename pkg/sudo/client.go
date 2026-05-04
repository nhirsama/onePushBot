package sudo

import (
	"context"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
)

type Client interface {
	SendGroupText(ctx context.Context, groupID string, text string) error
}

func clientFromEvent(source platformclient.Source, event base.Event) (Client, bool) {
	return platformclient.ForEvent[Client](source, event)
}

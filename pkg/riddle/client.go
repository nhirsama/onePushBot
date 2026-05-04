package riddle

import (
	"context"

	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/pkg/platform_client"
)

type Client interface {
	GetGroupMemberInfo(ctx context.Context, groupID string, userID string, noCache bool) (*base.GroupMemberInfo, error)
}

func clientFromEvent(source platformclient.Source, event base.Event) (Client, bool) {
	return platformclient.ForEvent[Client](source, event)
}

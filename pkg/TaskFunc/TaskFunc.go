package TaskFunc

import "github.com/nhirsama/onePushBot/api"

type Config struct {
	TaskFunc   func(message *api.WebSocketMessage)
	ModuleName string
}

var ModuleList []Config = make([]Config, 0)

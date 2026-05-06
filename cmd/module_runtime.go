package cmd

import (
	"github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/pkg/modules"
	"github.com/nhirsama/onePushBot/pkg/plat"
	"github.com/spf13/viper"
)

func newModuleRuntime(hub platform.Hub) (modules.Runtime, error) {
	return modules.NewRuntime(modules.Dependencies{
		Plat: plat.New(hub),
	}, viper.GetStringSlice("modules"))
}

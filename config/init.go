package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nhirsama/onePushBot/api"
	"github.com/nhirsama/onePushBot/pkg/TaskFunc"
	"github.com/spf13/viper"
)

func init() {
	const configPath = "./data"
	const configFile = configPath + "/config.yaml"
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			fmt.Println("正在初始化配置文件 :", configFile)
			// 确保目录存在
			if err := os.MkdirAll(configPath, 0755); err != nil {
				log.Fatal(err)
			}
			var input string

			fmt.Println("请输入api地址：")
			fmt.Scanln(&input)
			viper.Set("apiUrl", input)
			log.Printf("已设置apiUrl为 %s\n", input)

			fmt.Println("请输入token：")
			fmt.Scanln(&input)
			viper.Set("token", input)
			log.Printf("已设置token为 %s\n", input)

			fmt.Println("请输入管理员QQ号：")
			fmt.Scanln(&input)
			viper.Set("master", input)
			log.Printf("已设置master为 %s\n", input)

			if err := viper.WriteConfig(); err != nil {
				if err := viper.WriteConfigAs(configFile); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	fmt.Printf("%+v\n", viper.AllSettings())
	ReadConfig()

	TaskFunc.ModuleList = append(TaskFunc.ModuleList, TaskFunc.Config{TaskFunc: func(w *api.WebSocketMessage) {
		w.Bus.SubscribeAsync("groupMessage", func(msg *api.MessageStruct) {
			if strings.Contains(string(msg.Message), "\"type\":\"text\"") {
				DB.UpdateUser(msg.UserId, msg.RawMessage)
			}
		}, false)
	}, ModuleName: "groupMessageTokenDB"})
}

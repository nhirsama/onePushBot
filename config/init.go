package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

func init() {
	const configPath = "./data"
	const configFile = configPath + "/config.yaml"
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	setDefaults()

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
			inputInt, err := strconv.Atoi(input)
			if err == nil {
				viper.Set("master", inputInt)
				log.Printf("已设置master为 %s\n", input)
			}

			if err := viper.WriteConfig(); err != nil {
				if err := viper.WriteConfigAs(configFile); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	fmt.Printf("%+v\n", viper.AllSettings())
	ReadConfig()
}

func setDefaults() {
	// 平台层默认走 NapCat，保持旧配置最小可用。
	viper.SetDefault("platform", "qq")
	// 路由层默认缓冲只影响内部 fan-out，不改变平台总线语义。
	viper.SetDefault("router.broker_buffer", 128)
	// 飞书是 webhook 型平台，需要由 cmd 挂载 HTTP 入口。
	viper.SetDefault("feishu.http_addr", ":8080")
	viper.SetDefault("feishu.webhook_path", "/feishu/events")
}

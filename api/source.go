package api

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/asaskevich/EventBus"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type WebSocketMessage struct {
	mu                 sync.RWMutex
	conn               *websocket.Conn
	url                string
	WriteChan          chan commonRequest
	readGoroutineClose chan struct{}
	heartbeat          chan struct{}
	pubSub             *gochannel.GoChannel
	responseMap        sync.Map
	Bus                EventBus.Bus
}

func init() {
	viper.Reset()
	const configFile = "./data/config.yaml"
	const configPath = "./data"
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./data")

	//viper.SetConfigName(configFile)
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

	err := viper.UnmarshalKey("autoSetMsgEmojiLikeSet", &autoSetMsgEmojiLikeSet)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("%+v\n", viper.AllSettings())

}

func NewWebSocketMessage(url string) *WebSocketMessage {
	var w WebSocketMessage
	w.url = url
	w.reLogin()
	w.WriteChan = make(chan commonRequest, 100)
	w.heartbeat = make(chan struct{}, 1)
	//消息总线
	w.Bus = EventBus.New()
	return &w
}

func (w *WebSocketMessage) Close() {
	err := w.conn.Close()
	if err != nil {
		log.Println(err)
		return
	}
}

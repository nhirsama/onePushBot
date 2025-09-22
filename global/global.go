package global

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

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

}

// StatusStruct 心跳状态
type StatusStruct struct {
	Online bool `json:"online"`
	Good   bool `json:"good"`
}

// MessageData 消息内容结构体
type MessageData struct {
	Text string `json:"text"`
}

// MessageElement 消息元素（可能有多种类型，例如 text、image、emoji 等）
type MessageElement struct {
	Type string      `json:"type"`
	Data MessageData `json:"data"`
}

// Sender 发送者信息
type Sender struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
	Role     string `json:"role"`
}

// Message 通用消息结构体（兼容心跳、群消息等）
type Message struct {
	Time          int64            `json:"time"`
	SelfId        int64            `json:"self_id"`
	PostType      string           `json:"post_type"`
	MetaEventType string           `json:"meta_event_type,omitempty"`
	Status        *StatusStruct    `json:"status,omitempty"`
	Interval      int64            `json:"interval,omitempty"`
	MessageId     int64            `json:"message_id,omitempty"`
	MessageSeq    int64            `json:"message_seq,omitempty"`
	RealId        int64            `json:"real_id,omitempty"`
	RealSeq       string           `json:"real_seq,omitempty"`
	MessageType   string           `json:"message_type,omitempty"`
	Sender        *Sender          `json:"sender,omitempty"`
	RawMessage    string           `json:"raw_message,omitempty"`
	Font          int              `json:"font,omitempty"`
	SubType       string           `json:"sub_type,omitempty"`
	Message       []MessageElement `json:"message,omitempty"`
	MessageFormat string           `json:"message_format,omitempty"`
	GroupId       int64            `json:"group_id,omitempty"`
	GroupName     string           `json:"group_name,omitempty"`
	UserId        int64            `json:"user_id,omitempty"`
}

type ApiResponse struct {
	Status  string `json:"status"`
	RetCode int64  `json:"retcode"`
	Data    struct {
		Result int64  `json:"result"`
		ErrMsg string `json:"errMsg"`
	}
	Message string `json:"message"`
	Wording string `json:"wording"`
	Echo    string `json:"echo"`
}

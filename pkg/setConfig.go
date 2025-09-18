package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Deprecated: 请使用 viper.GitString 代替 Config。
// Config 已被废弃
type Config struct {
	ApiUrl string `json:"apiUrl"`
	Token  string `json:"token"`
	Master string `json:"master"`
}

// Deprecated: 请使用 viper.GitString 代替 SetConfig。
// SetConfig 已被废弃
func SetConfig() error {
	var filename string = "./data/config.json"
	var configMap map[string]interface{}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		configMap = make(map[string]interface{})
	} else {
		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &configMap); err != nil {
				configMap = make(map[string]interface{})
			}
		} else {
			configMap = make(map[string]interface{})
		}
	}
	fmt.Println("请输入api地址：")
	var inputString string
	fmt.Scanln(&inputString)
	configMap["apiUrl"] = inputString
	log.Printf("已设置apiUrl为%s\n", inputString)

	fmt.Println("请输入token：")
	fmt.Scanln(&inputString)
	configMap["token"] = inputString
	log.Printf("已设置token为%s\n", inputString)

	fmt.Println("请输入管理员QQ号：")
	fmt.Scanln(&inputString)
	configMap["master"] = inputString
	log.Printf("已设置master为%s\n", inputString)

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// 4. 重新写入文件
	newData, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filename, newData, 0644); err != nil {
		return err
	}

	return nil
}

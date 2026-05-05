package config

import "os"

const defaultDBPath = "./data/onepushbot.db"

// DBPath 来自环境变量，未设置时使用本地默认路径。
func DBPath() string {
	if value := os.Getenv("ONEPUSHBOT_DB_PATH"); value != "" {
		return value
	}
	return defaultDBPath
}

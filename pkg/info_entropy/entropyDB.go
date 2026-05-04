package infoEntropy

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type UserTokenInfo struct {
	Token      map[string]int
	TokenTotal int64 `json:"TokenTotal"`
}
type EntropyDB struct {
	mu        sync.RWMutex
	UserToken map[int64]*UserTokenInfo `json:"UserToken"`
}

const dbPath = "./data/entropy.json"

// LoadDB 载入数据库（若不存在则创建）
func LoadDB() *EntropyDB {
	db := &EntropyDB{UserToken: make(map[int64]*UserTokenInfo)}
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return db
	}

	data, err := os.ReadFile(dbPath)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return db
	}

	err = json.Unmarshal(data, db)
	if err != nil {
		return nil
	}
	return db
}

// Save 保存数据库
func (db *EntropyDB) Save() {
	db.mu.Lock()
	defer db.mu.Unlock()
	data, _ := json.MarshalIndent(db, "", "  ")
	os.WriteFile(dbPath, data, 0644)
}

func (db *EntropyDB) NewUserToken() *UserTokenInfo {
	var UserTokenInfo UserTokenInfo
	UserTokenInfo.Token = make(map[string]int)
	UserTokenInfo.TokenTotal = 0
	return &UserTokenInfo
}

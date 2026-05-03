package config

type legacyAuth interface {
	Authenticate(message []byte) (string, bool)
}

type legacyEntropyDB interface {
	UpdateUser(userID int64, text string)
	Save()
}

type noopAuth struct{}

func (noopAuth) Authenticate([]byte) (string, bool) {
	return "", false
}

type noopEntropyDB struct{}

func (noopEntropyDB) UpdateUser(int64, string) {}

func (noopEntropyDB) Save() {}

var AutoSetMsgEmojiLikeSet map[string][]int
var SelfId int64
var ApiKey string

var WebSocketUrl string
var Auth legacyAuth = noopAuth{}
var DB legacyEntropyDB = noopEntropyDB{}

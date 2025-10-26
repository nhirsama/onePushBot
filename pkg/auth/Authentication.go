package auth

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
)

type Auth struct {
	pubs    map[string]struct{}
	queue   []*VerifyResult
	mu      sync.RWMutex
	keyRing openpgp.EntityList
}

func NewAuth() *Auth {
	Auth := &Auth{
		pubs: make(map[string]struct{}),
	}
	return Auth
}

// Authenticate 传入pgp信息，若签名认证通过则返回签名信息，否则为空
func (a *Auth) Authenticate(message []byte) (string, bool) {

	re := a.VerifyPGPSignature(message)
	if !re.Valid {
		log.Println("invalid signature", re.Err)
		return "", false
	} else {
		// 丢弃五分钟前的所有签名
		if time.Now().Unix()-re.SignedAt > 300 {
			log.Println("signature expired")
			return "", false
		}
		// 使用短hash和时间戳拼接作为key
		hashkey := re.Hash + strconv.FormatInt(re.SignedAt, 10)

		a.mu.Lock()
		// 使用队列和哈希表查重和去除超时记录
		for len(a.queue) > 0 && time.Now().Unix()-a.queue[0].SignedAt > 300 {
			newHashKey := a.queue[0].Hash + strconv.FormatInt(a.queue[0].SignedAt, 10)
			delete(a.pubs, newHashKey)
			a.queue = a.queue[1:]
		}
		if _, ok := a.pubs[hashkey]; ok {
			return "", false
		}
		a.queue = append(a.queue, &re)
		a.pubs[hashkey] = struct{}{}
		defer a.mu.Unlock()
		return string(re.Message), true
	}
}

func (a *Auth) AddPublicKey(pubKey string) {
	keyRing, err := openpgp.ReadArmoredKeyRing(strings.NewReader(pubKey))
	if err != nil {
		log.Printf("警告：failed to read keyring from hardcoded data: %s", err)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.keyRing = append(a.keyRing, keyRing[0])
}

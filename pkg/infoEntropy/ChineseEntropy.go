package infoEntropy

import (
	"math"

	"github.com/yanyiwu/gojieba"
)

func (e *EntropyDB) UpdateUser(userId int64, text string) {
	x := gojieba.NewJieba()
	defer x.Free()

	// 精确模式分词
	words := x.Cut(text, true)
	if _, ok := e.UserToken[userId]; !ok {
		e.UserToken[userId] = e.NewUserToken()
	}
	for _, w := range words {
		if w == "" {
			continue
		}
		e.UserToken[userId].Token[w]++
		e.UserToken[userId].TokenTotal++
	}
}

func (e *EntropyDB) GetEntropy(userId int64) float64 {
	if _, ok := e.UserToken[userId]; ok {
		var H float64
		for _, c := range e.UserToken[userId].Token {
			p := float64(c) / float64(e.UserToken[userId].TokenTotal)
			H -= p * math.Log2(p)
		}
		return H
	}
	return 0.0
}

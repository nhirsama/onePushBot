package api

import (
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var ResponseMap sync.Map

func init() {
	err := viper.UnmarshalKey("autoSetMsgEmojiLikeSet", &autoSetMsgEmojiLikeSet)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("%+v\n", viper.AllSettings())
}

package global

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var C *websocket.Conn
var Heartbeat int64
var ResponseMap sync.Map

func init() {
	Heartbeat = time.Now().Unix()
}

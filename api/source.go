package api

import (
	"log"
	"sync"

	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/asaskevich/EventBus"
	"github.com/gorilla/websocket"
	_ "github.com/nhirsama/onePushBot/config"
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

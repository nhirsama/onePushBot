package api

import (
	"github.com/gorilla/websocket"
)

func Send_like(user_id int, times int, echo string, c *websocket.Conn) error {
	type params struct {
		User_id int    `json:"user_id"`
		Times   int    `json:"times"`
		Echo    string `json:"echo"`
	}
	type send_like struct {
		Action string `json:"action"`
		Parmes params `json:"params"`
	}

	body := send_like{"send_like", params{user_id, times, echo}}
	err := c.WriteJSON(body)
	return err
}

/*
{
  "action": "send_like",
  "params":{
    "user_id" : xxxxxx,
    "times": 10
  }
}
*/

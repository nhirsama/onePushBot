package api

import "github.com/gorilla/websocket"

func Send_private_msg_text(user_id int, text string, echo string, c *websocket.Conn) error {
	type data struct {
		Text string `json:"text"`
	}
	type message struct {
		Type string `json:"type"`
		Data data   `json:"data"`
	}
	type params struct {
		User_id int       `json:"user_id"`
		Echo    string    `json:"echo"`
		Message []message `json:"message"`
	}
	type send_private_msg struct {
		Action string `json:"action"`
		Parmes params `json:"params"`
	}
	body := send_private_msg{"send_private_msg",
		params{user_id, echo, make([]message, 1)}}
	body.Parmes.Message[0].Type = "text"
	body.Parmes.Message[0].Data.Text = text
	err := c.WriteJSON(body)
	return err
}

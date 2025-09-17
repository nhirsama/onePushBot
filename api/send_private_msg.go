package api

func Send_private_msg(user_id int) error {
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
	body := send_private_msg{"send_private_msg", params{}}
}

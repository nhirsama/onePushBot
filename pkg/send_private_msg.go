package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type MessageData struct {
	Text string `json:"text"`
}

type Message struct {
	Type string      `json:"type"`
	Data MessageData `json:"data"`
}

type Payload struct {
	UserID  string    `json:"user_id"`
	Message []Message `json:"message"`
}

func SendPrivateMsg(config Config, text string) error {
	apiName, _ := url.Parse("send_private_msg")
	baseUrl, _ := url.Parse(config.ApiUrl)
	baseUrl = baseUrl.ResolveReference(apiName)
	method := "POST"

	payload := Payload{
		UserID: config.Master,
		Message: []Message{
			{
				Type: "text",
				Data: MessageData{
					Text: text,
				},
			},
		},
	}

	client := &http.Client{}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, baseUrl.String(), bytes.NewBuffer(jsonPayload))

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	log.Println(string(body))
	return nil
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/nhirsama/onePushBot/pkg"
)

func main() {
	data, err := os.ReadFile("data/config.json")
	if err != nil {
		log.Println(err)
		err = pkg.SetConfig()
		if err != nil {
			log.Fatal(err)
		}
		data, err = os.ReadFile("data/config.json")
		if err != nil {
			log.Fatal(err)
		}
	}
	var config pkg.Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}

	var inputString string
	fmt.Scanln(&inputString)
	err = pkg.SendPrivateMsg(config, inputString)
	if err != nil {
		log.Println(err)
	}
}

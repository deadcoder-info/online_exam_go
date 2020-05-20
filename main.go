package main

import (
	"encoding/json"
	"log"
	"online_exam_go/databaselayer"
	"online_exam_go/examwebportal"
	"os"
	"fmt"
)

type configration struct {
	ServerAddress string `json:"webserver"`
}

func main() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("err")
	}
	config := new(configration)
	json.NewDecoder(file).Decode(config)
	db := databaselayer.NewDataBase()
	examwebportal.RunWebPortal(config.ServerAddress, db)
	fmt.Println("Server is running on: " + config.ServerAddress)
}

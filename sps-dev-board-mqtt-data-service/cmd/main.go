package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sps-dev-board-mqtt-data-service/pkg/data-service"
)

func main() {
	fmt.Println("Hello devboard with S7-1200 PLC")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mqttHost := os.Getenv("MQTT_BROKER_HOST")


	data_service.StartDataService(mqttHost)
}

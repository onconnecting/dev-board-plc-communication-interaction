package data_service

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"time"
)

type DataService struct {
	mqttClient                      mqtt.Client
	lastReceivedDistanceSensorInput uint
	currentInductiveSensorOutput    uint
	currentDistanceSensorOutput     uint

	manualMode      bool
	manualDistance  uint
	manualInductive uint
}

var globalDataService DataService

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

	switch msg.Topic() {
	case "devBoard/ioLinkMaster/port/7":
		handleTemperatureTopic(client, msg)
	case "devBoard/ioLinkMaster/port/2/pdi":
		handleRealDistanceSensorTopic(client, msg)
	case "devBoard/ioLinkMaster/port/1/pdi":
		handleInductiveSensorTopic(client, msg)
	default:
		fmt.Printf("Received from unknown topic %s\n", msg.Topic())
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func StartDataService(broker string) {
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.SetDefaultPublishHandler(messagePubHandler)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscribe to the topic
	if token := client.Subscribe("devBoard/ioLinkMaster/port/2/pdi", 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	// Subscribe to the topic
	if token := client.Subscribe("devBoard/ioLinkMaster/port/1/pdi", 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	dataService := DataService{
		mqttClient: client,
	}

	globalDataService = dataService

	setupWebserver()

	for {
		distance := uint(10)

		globalDataService.currentDistanceSensorOutput = distance

		publishDataToPLC()

		time.Sleep(2 * time.Second)
	}
}

func setupWebserver() {
	r := gin.Default()

	// Serve HTML files
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API endpoint to handle input changes
	r.POST("/update-settings", func(c *gin.Context) {
		type Settings struct {
			ManualMode      bool   `json:"manualMode"`
			InductiveSensor bool   `json:"inductiveSensor"`
			DistanceSensor  string `json:"distanceSensor"`
		}

		var settings Settings

		// Bind JSON to struct
		if err := c.BindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		globalDataService.manualMode = settings.ManualMode

		if settings.InductiveSensor {
			globalDataService.manualInductive = 1
		} else {
			globalDataService.manualInductive = 0
		}

		distanceAsInt, _ := strconv.Atoi(settings.DistanceSensor)

		globalDataService.manualDistance = uint(distanceAsInt)

		publishDataToPLC()

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	r.Run(":8080")
}


func handleDistanceSensorTopic(client mqtt.Client, msg mqtt.Message) {
	type DistanceSensorMessage struct {
		Uint uint `json:"uint"`
	}

	var parsedMessage DistanceSensorMessage

	json.Unmarshal(msg.Payload(), &parsedMessage)

	globalDataService.lastReceivedDistanceSensorInput = parsedMessage.Uint
}

func handleRealDistanceSensorTopic(client mqtt.Client, msg mqtt.Message) {
	type RealDistanceSensorMessage struct {
		Vpdt struct {
			Distance uint `json:"Distance"`
		} `json:"V_PdT"`
	}

	var parsedMessage RealDistanceSensorMessage

	json.Unmarshal(msg.Payload(), &parsedMessage)

	globalDataService.currentDistanceSensorOutput = parsedMessage.Vpdt.Distance;

	publishDataToPLC()
}

func handleInductiveSensorTopic(client mqtt.Client, msg mqtt.Message) {
	type InductiveSensorMessage struct {
		Vpdt struct {
			SSC bool `json:"SSC1"`
		} `json:"V_PdT"`
	}

	var parsedMessage InductiveSensorMessage

	json.Unmarshal(msg.Payload(), &parsedMessage)

	if parsedMessage.Vpdt.SSC {
		globalDataService.currentInductiveSensorOutput = 1
	} else {
		globalDataService.currentInductiveSensorOutput = 0
	}

	publishDataToPLC()
}

func handleTemperatureTopic(client mqtt.Client, msg mqtt.Message) {
	temperature := 1
	// Modify the message
	newMessage := fmt.Sprintf(
		`{"port":7,"valid":1,"uint":15990528,"V_PdT":{"Temperature":%d,"OUT2":false,"OUT1":false},"raw":[0,243,255,0]}`,
		temperature,
	)

	// Publish to a new topic
	token := client.Publish("devBoard/plcS71200", 0, false, newMessage)
	token.Wait()
}

func publishDataToPLC() {
	var paddedDistance, paddedInductive = "", ""

	if globalDataService.manualMode {
		paddedDistance = fmt.Sprintf("%03d", globalDataService.manualDistance)
		paddedInductive = fmt.Sprintf("%01d", globalDataService.manualInductive)
	} else {
		paddedDistance = fmt.Sprintf("%03d", globalDataService.currentDistanceSensorOutput)
		paddedInductive = fmt.Sprintf("%01d", globalDataService.currentInductiveSensorOutput)
	}

	fmt.Printf("Got distance of %s\n", paddedDistance)
	fmt.Printf("Got indutive of %s\n", paddedInductive)

	token := globalDataService.mqttClient.Publish("devBoard/plcS71200", 0, false, paddedDistance+";"+paddedInductive)
	token.Wait()
}

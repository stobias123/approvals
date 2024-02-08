package main

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/stobias123/slack_button/http_handlers"
	"github.com/stobias123/slack_button/managers"
)

type ApprovalPayload struct {
	Approved   bool   `json:"approved"`
	ApprovalID string `json:"approval_id"`
}

type Server struct {
	Router *gin.Engine
}

func NewServer(sm *managers.SlackManager, am *managers.ApprovalManager) *Server {
	router := gin.Default()

	// MQTT setup
	opts := mqtt.NewClientOptions().AddBroker("tcp://mqtturl:1883").SetClientID("go_mqtt_client")
	opts.SetUsername("mqttuser")
	opts.SetPassword("mqttPass")
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}

	approvalsController := http_handlers.NewApprovalsController(mqttClient, sm, am)
	approvalsController.RegisterHandlers(router)

	return &Server{
		Router: router,
	}
}

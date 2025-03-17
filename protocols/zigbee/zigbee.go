package zigbee

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

const (
	mqttBroker       = "tcp://localhost:1883"
	zigbee2mqttTopic = "zigbee2mqtt/#"
)

var (
	deviceStates = make(map[string]interface{}) // Zigbee device states
	mu           = sync.Mutex{}
	mqttClient   mqtt.Client
)

// MQTT message handler
func messageHandler(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := string(msg.Payload())

	log.Printf("📡 Zigbee | Topic: %s | Payload: %s", topic, payload)

	deviceID := strings.TrimPrefix(topic, "zigbee2mqtt/")
	if strings.HasSuffix(deviceID, "/set") || strings.HasPrefix(deviceID, "bridge/") {
		return
	}

	var state map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &state); err == nil {
		mu.Lock()
		deviceStates[deviceID] = state
		mu.Unlock()
	}
}

// Initialize Zigbee MQTT
func InitMQTT() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttBroker)
	opts.SetClientID("zigbee2mqtt-go-api")
	opts.SetDefaultPublishHandler(messageHandler)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Zigbee MQTT Connection Failed: %v", token.Error())
	}

	mqttClient.Subscribe(zigbee2mqttTopic, 0, nil)
	log.Println("✅ Subscribed to Zigbee2MQTT")

	// Request device states
	mqttClient.Publish("zigbee2mqtt/bridge/request/device_state", 0, false, "{}")
}

// Get all Zigbee devices
func GetDevices() map[string]interface{} {
	mu.Lock()
	defer mu.Unlock()

	filteredDevices := make(map[string]interface{})
	for deviceID, data := range deviceStates {
		filteredDevices["zigbee_"+deviceID] = data
	}
	return filteredDevices
}

// Send Zigbee command
func SendDeviceCommand(c *gin.Context) {
	deviceID := c.Param("device_id")
	var command map[string]interface{}

	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	commandJSON, _ := json.Marshal(command)
	mqttClient.Publish(fmt.Sprintf("zigbee2mqtt/%s/set", deviceID), 0, false, commandJSON)
	c.JSON(http.StatusOK, gin.H{"status": "Command sent", "command": command})
}

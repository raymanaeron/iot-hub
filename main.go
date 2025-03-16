package main

import (
        "encoding/json"
        "fmt"
        "log"
        "net/http"
        "sync"
        "strings"

        mqtt "github.com/eclipse/paho.mqtt.golang"
        "github.com/gin-gonic/gin"
)

// MQTT Broker & Zigbee2MQTT Topic
const (
        mqttBroker       = "tcp://localhost:1883"
        zigbee2mqttTopic = "zigbee2mqtt/#"
)

var (
        deviceStates = make(map[string]interface{}) // Stores device state
        mu          = sync.Mutex{}                  // Mutex to prevent race conditions
        mqttClient  mqtt.Client                      // MQTT client
)

// Callback for MQTT messages
func messageHandler(client mqtt.Client, msg mqtt.Message) {
    topic := msg.Topic()
    payload := string(msg.Payload())

    log.Printf("📡 Received MQTT Message | Topic: %s | Payload: %s", topic, payload)

    deviceID := topic[len("zigbee2mqtt/"):] // Extract device ID

    if len(deviceID) > 4 && deviceID[len(deviceID)-4:] == "/set" {
        return // Ignore "set" messages
    }

    var state map[string]interface{}
    if err := json.Unmarshal(msg.Payload(), &state); err == nil {
        mu.Lock()
        deviceStates[deviceID] = state
        mu.Unlock()
        return
    }

    log.Printf("⚠️ Failed to parse MQTT message: %s", payload)
}

// Initialize MQTT connection
func initMQTT() {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(mqttBroker)
    opts.SetClientID("zigbee2mqtt-go-api")
    opts.SetDefaultPublishHandler(messageHandler)
    opts.SetAutoReconnect(true)
    opts.SetCleanSession(true)

    mqttClient = mqtt.NewClient(opts)
    if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
        log.Fatalf("❌ Failed to connect to MQTT broker: %v", token.Error())
    }

    mqttClient.Subscribe(zigbee2mqttTopic, 0, nil)
    log.Println("✅ Subscribed to Zigbee2MQTT messages")

    // 🔥 Force all devices to report their current state on API startup
    mqttClient.Publish("zigbee2mqtt/bridge/request/device_state", 0, false, "{}")
    log.Println("🔄 Requested all device states from Zigbee2MQTT")
}

// **GET /devices - List all Zigbee devices**
func listDevices(c *gin.Context) {
    mu.Lock()
    defer mu.Unlock()

    filteredDevices := make(map[string]interface{})

    for deviceID, data := range deviceStates {
        if !strings.HasPrefix(deviceID, "bridge/") { // Ignore system messages
            filteredDevices[deviceID] = data // Keep full device data
        }
    }

    c.JSON(http.StatusOK, filteredDevices)
}

// **GET /devices/:device_id - Get details of a Zigbee device**
func getDevice(c *gin.Context) {
        deviceID := c.Param("device_id")

        mu.Lock()
        state, exists := deviceStates[deviceID]
        mu.Unlock()

        if exists {
                c.JSON(http.StatusOK, state)
        } else {
                c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
        }
}

// **GET /devices/discover - Start discovering new Zigbee devices**
func discoverDevices(c *gin.Context) {
        mqttClient.Publish("zigbee2mqtt/bridge/request/permit_join", 0, false, "true")
        c.JSON(http.StatusOK, gin.H{"status": "Discovery started"})
}

// **POST /devices/:device_id/set - Send a command to a device**
func sendCommand(c *gin.Context) {
        deviceID := c.Param("device_id")
        var command map[string]interface{}

        if err := c.ShouldBindJSON(&command); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
                return
        }

        commandJSON, _ := json.Marshal(command)
        mqttClient.Publish(fmt.Sprintf("zigbee2mqtt/%s/set", deviceID), 0, false, commandJSON)
        c.JSON(http.StatusOK, gin.H{"status": "Command sent", "command": command})
}

func main() {
        // Initialize MQTT connection
        initMQTT()

        // Setup API routes
        r := gin.Default()
        r.GET("/devices", listDevices)
        r.GET("/devices/:device_id", getDevice)
        r.GET("/devices/discover", discoverDevices)
        r.POST("/devices/:device_id/set", sendCommand)

        // Start HTTP server
        log.Println("🚀 Zigbee2MQTT API running at http://localhost:3000")
        r.Run(":3000")
}


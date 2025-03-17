package main

import (
	"log"
	"net/http"

	"iothub/protocols/wifi"
	"iothub/protocols/zigbee"

	"github.com/gin-gonic/gin"
)

// Get all devices from Zigbee, Z-Wave, Matter, WiFi, BLE
func GetAllDevices(c *gin.Context) {
	allDevices := make(map[string]interface{})

	// Get devices from each protocol
	zigbeeDevices := zigbee.GetDevices()

	// Get WiFi devices
	wifiDevices := wifi.GetDevices()

	// Merge all results
	for id, data := range zigbeeDevices {
		allDevices[id] = data
	}

	for id, data := range wifiDevices {
		allDevices["wifi-"+id] = data
	}

	/*
		for id, data := range zwaveDevices {
			allDevices[id] = data
		}
		for id, data := range matterDevices {
			allDevices[id] = data
		}
		for id, data := range wifiDevices {
			allDevices[id] = data
		}
		for id, data := range bleDevices {
			allDevices[id] = data
		}
	*/

	c.JSON(http.StatusOK, allDevices)
}

// ** GET /devices/wifi/discover - Scan for WiFi devices **
func discoverWiFiDevices(c *gin.Context) {
	devices := wifi.ScanDevices()
	c.JSON(http.StatusOK, devices)
}

// ** POST /devices/wifi/:ip/set - Send a command to a WiFi device **
func sendWiFiCommand(c *gin.Context) {
	ip := c.Param("ip")
	var command map[string]interface{}

	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	err := wifi.SendCommand(ip, command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Command sent", "command": command})
}

func main() {
	// Initialize all protocol handlers
	zigbee.InitMQTT()
	//zwave.InitMQTT()
	//matter.Init()
	//wifi.Init()
	//ble.Init()

	// Setup API routes
	r := gin.Default()
	r.GET("/devices", GetAllDevices)
	r.POST("/devices/:device_id/set", zigbee.SendDeviceCommand) // Unified command handler
	r.GET("/devices/wifi/discover", discoverWiFiDevices)
	r.POST("/devices/wifi/:ip/set", sendWiFiCommand)

	// Start HTTP server
	log.Println("🚀 IoT Hub API running at http://localhost:3000")
	r.Run(":3000")
}

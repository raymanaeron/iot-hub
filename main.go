package main

import (
	"log"
	"net/http"

	"iothub/protocols/zigbee"

	"github.com/gin-gonic/gin"
)

// Get all devices from Zigbee, Z-Wave, Matter, WiFi, BLE
func GetAllDevices(c *gin.Context) {
	allDevices := make(map[string]interface{})

	// Get devices from each protocol
	zigbeeDevices := zigbee.GetDevices()

	// Merge all results
	for id, data := range zigbeeDevices {
		allDevices[id] = data
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

	// Start HTTP server
	log.Println("🚀 IoT Hub API running at http://localhost:3000")
	r.Run(":3000")
}

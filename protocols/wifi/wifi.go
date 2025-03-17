package wifi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

// DeviceState represents a WiFi IoT device
type DeviceState struct {
	IP    string                 `json:"ip"`
	Mac   string                 `json:"mac"`
	State map[string]interface{} `json:"state,omitempty"`
}

var (
	deviceStates = make(map[string]DeviceState) // Stores WiFi devices
	mu           sync.Mutex                     // Mutex for concurrency
)

// **Scan for WiFi devices (Using `arp -a`)**
func ScanDevices() map[string]DeviceState {
	mu.Lock()
	defer mu.Unlock()

	devices := make(map[string]DeviceState)

	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️ Failed to scan WiFi devices: %v", err)
		return devices
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ip := strings.Trim(fields[0], "()")
			mac := fields[1]
			devices[ip] = DeviceState{IP: ip, Mac: mac}
		}
	}

	deviceStates = devices
	return devices
}

// **Get all known WiFi devices**
func GetDevices() map[string]DeviceState {
	mu.Lock()
	defer mu.Unlock()
	return deviceStates
}

// **Send a command to a WiFi IoT device**
func SendCommand(ip string, command map[string]interface{}) error {
	jsonData, err := json.Marshal(command)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/api", ip) // Assuming the WiFi device uses an HTTP API
	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("device %s responded with status: %d", ip, resp.StatusCode)
	}

	return nil
}

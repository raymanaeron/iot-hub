package wifi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
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
	deviceStates = make(map[string]DeviceState) // Stores discovered WiFi devices
	mu           sync.Mutex                     // Mutex for safe concurrent access
)

// **Scan for WiFi devices (Using `arp -a`)**
func ScanDevices() map[string]DeviceState {
	mu.Lock()
	defer mu.Unlock()

	devices := make(map[string]DeviceState)

	// Run `arp -a` to list network devices
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️ Failed to scan WiFi devices: %v", err)
		return devices
	}

	lines := strings.Split(string(output), "\n")

	// Regex pattern to extract IP & MAC addresses from arp output
	re := regexp.MustCompile(`\((\d+\.\d+\.\d+\.\d+)\)\s+at\s+([0-9a-fA-F:-]+)`)

	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		if len(match) == 3 {
			ip := match[1]  // Extract IP address
			mac := match[2] // Extract MAC address
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
		return fmt.Errorf("failed to encode command: %v", err)
	}

	url := fmt.Sprintf("http://%s/api", ip) // Assuming WiFi device has an HTTP API
	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to send command to device %s: %v", ip, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("device %s responded with status: %d", ip, resp.StatusCode)
	}

	return nil
}

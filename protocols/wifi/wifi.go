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

// **WiFiDevice represents all available properties of a WiFi IoT device**
type WiFiDevice struct {
	IP       string                 `json:"ip"`
	MAC      string                 `json:"mac"`
	Hostname string                 `json:"hostname,omitempty"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

var (
	deviceStates = make(map[string]WiFiDevice) // Stores discovered WiFi devices
	mu           sync.Mutex                    // Mutex for concurrency
)

// **Scan for WiFi devices (Using `arp -a`)**
func ScanDevices() map[string]WiFiDevice {
	mu.Lock()
	defer mu.Unlock()

	devices := make(map[string]WiFiDevice)

	// Run `arp -a` to list network devices
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("⚠️ Failed to scan WiFi devices: %v", err)
		return devices
	}

	lines := strings.Split(string(output), "\n")

	// Regex pattern to extract IP, MAC, and hostname
	re := regexp.MustCompile(`(?P<hostname>\S+)?\s+\((?P<ip>\d+\.\d+\.\d+\.\d+)\)\s+at\s+(?P<mac>[0-9a-fA-F:-]+)`)

	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		if len(match) >= 4 {
			ip := match[2]
			mac := match[3]
			hostname := match[1]

			// Store full device data
			devices[ip] = WiFiDevice{
				IP:       ip,
				MAC:      mac,
				Hostname: hostname,
				Extra:    map[string]interface{}{}, // Placeholder for additional properties
			}
		}
	}

	deviceStates = devices
	return devices
}

// **Get all known WiFi devices**
func GetDevices() map[string]WiFiDevice {
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

	url := fmt.Sprintf("http://%s/api", ip) // Assuming the WiFi device has an HTTP API
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

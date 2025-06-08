package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"time"
)

// MonitorStatus represents response structure for monitor status
type MonitorStatus struct {
	Name           string `json:"name"`
	Endpoint       string `json:"endpoint"`
	Status         string `json:"status"`
	UnchangedCount int    `json:"unchangedCount,omitempty"`
}

// NewMonitorService initializes a new monitoring service
func NewMonitorService(monitors []*Monitor) *MonitorService {
	return &MonitorService{
		Monitors: monitors,
		stopChan: make(chan struct{}),
	}
}

// StartAll begins monitoring all configured endpoints
func (s *MonitorService) StartAll() {
	for _, monitor := range s.Monitors {
		go s.startMonitoring(monitor)
	}
}

// StopAll stops all monitoring activities
func (s *MonitorService) StopAll() {
	close(s.stopChan)
}

// GetAllStatus returns status information for all monitors
func (s *MonitorService) GetAllStatus() []MonitorStatus {
	result := make([]MonitorStatus, len(s.Monitors))
	for i, m := range s.Monitors {
		result[i] = MonitorStatus{
			Name:     m.Name,
			Endpoint: m.Endpoint,
			Status:   m.Status,
		}
	}
	return result
}

// GetStatus returns status information for a specific monitor
func (s *MonitorService) GetStatus(name string) *MonitorStatus {
	for _, m := range s.Monitors {
		if m.Name == name {
			return &MonitorStatus{
				Name:           m.Name,
				Endpoint:       m.Endpoint,
				Status:         m.Status,
				UnchangedCount: m.UnchangedCount,
			}
		}
	}
	return nil
}

// startMonitoring starts the monitoring process for a specific endpoint
func (s *MonitorService) startMonitoring(monitor *Monitor) {
	log.Printf("Starting monitor for %s (%s)", monitor.Name, monitor.Endpoint)

	// Perform initial check immediately
	s.checkEndpoint(monitor)

	ticker := time.NewTicker(time.Duration(monitor.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkEndpoint(monitor)
		case <-s.stopChan:
			log.Printf("Stopping monitor for %s", monitor.Name)
			return
		}
	}
}

// checkEndpoint performs a check on an endpoint and updates its status
func (s *MonitorService) checkEndpoint(monitor *Monitor) {
	log.Printf("Checking %s (%s)...", monitor.Name, monitor.Endpoint)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(monitor.Endpoint)
	if err != nil {
		log.Printf("Error checking %s: %v", monitor.Name, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response from %s: %v", monitor.Name, err)
		return
	}

	newHash := calculateHash(body)

	// First check - just store the hash
	if monitor.LastHash == "" {
		monitor.LastHash = newHash
		monitor.Status = "success"
		monitor.UnchangedCount = 0
		log.Printf("Initial check for %s completed. Hash: %s", monitor.Name, newHash)
		return
	}

	// Compare with previous hash
	if newHash == monitor.LastHash {
		// Content hasn't changed
		monitor.UnchangedCount++
		log.Printf("Monitor %s failed for %d of %d unchanged checks", monitor.Name, monitor.UnchangedCount, monitor.FailThreshold)

		if monitor.UnchangedCount >= monitor.FailThreshold {
			if monitor.Status != "fail" {
				log.Printf("Monitor %s marked as FAIL after %d unchanged checks", monitor.Name, monitor.UnchangedCount)
			}
			monitor.Status = "fail"
		}
	} else {
		// Content has changed
		if monitor.Status != "success" {
			log.Printf("Monitor %s content changed, marked as SUCCESS", monitor.Name)
		}
		monitor.Status = "success"
		monitor.UnchangedCount = 0
		monitor.LastHash = newHash
	}
}

// calculateHash computes a SHA-256 hash of the provided content
func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

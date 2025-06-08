package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

// Monitor represents a configured endpoint to monitor
type Monitor struct {
	Name          string `yaml:"name"`
	Endpoint      string `yaml:"endpoint"`
	CheckInterval int    `yaml:"checkInterval"` // in seconds
	FailThreshold int    `yaml:"failThreshold"` // number of unchanged checks before failure

	// Runtime state
	Status         string
	LastHash       string
	UnchangedCount int
}

// MonitorService manages all monitors and their state
type MonitorService struct {
	Monitors []*Monitor
	stopChan chan struct{}
}

func main() {
	log.Println("Starting Did-It-Change monitoring service...")

	// Load configuration
	config, err := loadConfig("config/monitors.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup monitoring service
	monitorService := NewMonitorService(config.Monitors)

	// Start monitoring
	monitorService.StartAll()

	// Setup API server
	router := gin.Default()
	setupRoutes(router, monitorService)

	// Start API server in a goroutine
	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("Server started on :8080")
	log.Printf("Monitoring %d endpoints", len(config.Monitors))

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	monitorService.StopAll()
}

func setupRoutes(router *gin.Engine, service *MonitorService) {
	router.GET("/api/monitors", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"count":    len(service.Monitors),
			"monitors": service.GetAllStatus(),
		})
	})

	router.GET("/api/monitors/:name", func(c *gin.Context) {
		name := c.Param("name")
		monitor := service.GetStatus(name)

		if monitor == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Monitor '%s' not found", name),
			})
			return
		}

		c.JSON(http.StatusOK, monitor)
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})
}

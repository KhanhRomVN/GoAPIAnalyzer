package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
}

type PingResponse struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health returns the health status of the application
func (h *HealthHandler) Health(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0", // This could be injected from build info
		Service:   "GoAPIAnalyzer",
	}

	c.JSON(http.StatusOK, response)
}

// Ping returns a simple pong response
func (h *HealthHandler) Ping(c *gin.Context) {
	response := PingResponse{
		Message:   "pong",
		Timestamp: time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Version string `json:"version"`
}

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// Health godoc
// @Summary Health check
// @Description Check if the API is running and healthy
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} handler.HealthResponse
// @Router /health [get]
func (h *SystemHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "healthy",
		Message: "API is running",
		Version: "1.0.0",
	})
}

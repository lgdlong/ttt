package handler

import (
	"api/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService domain.StatsService
}

func NewStatsHandler(statsService domain.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetAdminStats returns statistics for admin dashboard
// GET /api/v1/admin/stats
func (h *StatsHandler) GetAdminStats(c *gin.Context) {
	stats, err := h.statsService.GetAdminStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetModStats returns statistics for moderator dashboard
// GET /api/v1/mod/stats
func (h *StatsHandler) GetModStats(c *gin.Context) {
	stats, err := h.statsService.GetModStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get mod stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

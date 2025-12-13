package domain

import (
	"api/internal/dto"
)

type StatsService interface {
	GetAdminStats() (*dto.AdminStatsResponse, error)
	GetModStats() (*dto.ModStatsResponse, error)
}

package service

import (
	"api/internal/dto"
	"api/internal/repository"
)

type StatsService interface {
	GetAdminStats() (*dto.AdminStatsResponse, error)
	GetModStats() (*dto.ModStatsResponse, error)
}

type statsService struct {
	statsRepo repository.StatsRepository
}

func NewStatsService(statsRepo repository.StatsRepository) StatsService {
	return &statsService{
		statsRepo: statsRepo,
	}
}

// GetAdminStats returns statistics for admin dashboard
func (s *statsService) GetAdminStats() (*dto.AdminStatsResponse, error) {
	totalUsers, err := s.statsRepo.GetTotalUsers()
	if err != nil {
		return nil, err
	}

	activeUsers, err := s.statsRepo.GetActiveUsers()
	if err != nil {
		return nil, err
	}

	totalVideos, err := s.statsRepo.GetTotalVideos()
	if err != nil {
		return nil, err
	}

	totalTags, err := s.statsRepo.GetTotalTags()
	if err != nil {
		return nil, err
	}

	return &dto.AdminStatsResponse{
		TotalUsers:  totalUsers,
		ActiveUsers: activeUsers,
		TotalVideos: totalVideos,
		TotalTags:   totalTags,
	}, nil
}

// GetModStats returns statistics for moderator dashboard
func (s *statsService) GetModStats() (*dto.ModStatsResponse, error) {
	totalVideos, err := s.statsRepo.GetTotalVideos()
	if err != nil {
		return nil, err
	}

	totalTags, err := s.statsRepo.GetTotalTags()
	if err != nil {
		return nil, err
	}

	videosWithTranscript, err := s.statsRepo.GetVideosWithTranscript()
	if err != nil {
		return nil, err
	}

	videosAddedToday, err := s.statsRepo.GetVideosAddedToday()
	if err != nil {
		return nil, err
	}

	return &dto.ModStatsResponse{
		TotalVideos:          totalVideos,
		TotalTags:            totalTags,
		VideosWithTranscript: videosWithTranscript,
		VideosAddedToday:     videosAddedToday,
	}, nil
}

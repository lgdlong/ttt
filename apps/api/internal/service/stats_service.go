package service

import (
	"api/internal/domain"
	"api/internal/dto"
)

type statsService struct {
	statsRepo domain.StatsRepository
}

func NewStatsService(statsRepo domain.StatsRepository) domain.StatsService {
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

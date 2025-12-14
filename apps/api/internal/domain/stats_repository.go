package domain

type StatsRepository interface {
	GetTotalUsers() (int64, error)
	GetActiveUsers() (int64, error)
	GetTotalVideos() (int64, error)
	GetTotalTags() (int64, error)
	GetVideosWithTranscript() (int64, error)
	GetVideosAddedToday() (int64, error)
}

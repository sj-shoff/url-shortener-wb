package domain

import "time"

type Click struct {
	ID        int64
	URLID     int64
	UserAgent string
	IPAddress string
	ClickedAt time.Time
}

type AnalyticsReport struct {
	TotalClicks    int
	DailyStats     map[string]int
	UserAgentStats map[string]int
	Clicks         []Click
}

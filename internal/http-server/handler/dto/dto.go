package dto

type CreateShortURLRequest struct {
	URL    string `json:"url" validate:"required,url"`
	Custom string `json:"custom,omitempty" validate:"omitempty,min=3,max=20,alphanum"`
}

type CreateShortURLResponse struct {
	ShortURL string `json:"short_url"`
	Alias    string `json:"alias"`
}

type AnalyticsResponse struct {
	TotalClicks    int              `json:"total_clicks"`
	DailyStats     map[string]int   `json:"daily_stats"`
	MonthlyStats   map[string]int   `json:"monthly_stats"`
	UserAgentStats map[string]int   `json:"user_agent_stats"`
	Clicks         []ClickAnalytics `json:"clicks"`
}

type ClickAnalytics struct {
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
	ClickedAt string `json:"clicked_at"`
}

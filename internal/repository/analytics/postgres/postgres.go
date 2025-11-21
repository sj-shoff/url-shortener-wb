package postgres

import (
	"context"
	"url-shortener-wb/internal/domain"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type AnalyticsRepository struct {
	db      *dbpg.DB
	urlRepo URLRepository
	retries retry.Strategy
}

type URLRepository interface {
	GetByAlias(ctx context.Context, alias string) (*domain.URL, error)
}

func NewAnalyticsRepository(
	db *dbpg.DB,
	urlRepo URLRepository,
	retries retry.Strategy,
) *AnalyticsRepository {
	return &AnalyticsRepository{
		db:      db,
		urlRepo: urlRepo,
		retries: retries,
	}
}

func (r *AnalyticsRepository) RecordClick(ctx context.Context, click *domain.Click) error {
	_, err := r.db.ExecWithRetry(ctx, r.retries,
		`INSERT INTO clicks (url_id, user_agent, ip_address, clicked_at)
		VALUES ($1, $2, $3, $4)`,
		click.URLID, click.UserAgent, click.IPAddress, click.ClickedAt,
	)
	return err
}

func (r *AnalyticsRepository) GetAnalytics(ctx context.Context, alias string) (*domain.AnalyticsReport, error) {
	url, err := r.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryWithRetry(ctx, r.retries,
		`SELECT id, user_agent, ip_address, clicked_at
		FROM clicks WHERE url_id = $1
		ORDER BY clicked_at DESC
		LIMIT 100`, url.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []domain.Click
	for rows.Next() {
		var click domain.Click
		if err := rows.Scan(&click.ID, &click.UserAgent, &click.IPAddress, &click.ClickedAt); err != nil {
			return nil, err
		}
		click.URLID = url.ID
		clicks = append(clicks, click)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	report := &domain.AnalyticsReport{
		TotalClicks:    len(clicks),
		DailyStats:     make(map[string]int),
		UserAgentStats: make(map[string]int),
		Clicks:         clicks,
	}

	for _, click := range clicks {
		date := click.ClickedAt.Format("2006-01-02")
		report.DailyStats[date]++
		report.UserAgentStats[click.UserAgent]++
	}

	return report, nil
}

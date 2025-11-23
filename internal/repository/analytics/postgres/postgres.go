package postgres

import (
	"context"
	"errors"
	"fmt"

	"url-shortener-wb/internal/domain"
	repo "url-shortener-wb/internal/repository"

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
	if err != nil {
		return fmt.Errorf("failed to insert click: %w", err)
	}
	return nil
}

func (r *AnalyticsRepository) GetAnalytics(ctx context.Context, alias string) (*domain.AnalyticsReport, error) {
	url, err := r.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, fmt.Errorf("%w: url not found for alias %s", repo.ErrNotFound, alias)
		}
		return nil, fmt.Errorf("failed to get url by alias: %w", err)
	}

	rows, err := r.db.QueryWithRetry(ctx, r.retries,
		`SELECT id, user_agent, ip_address, clicked_at
		FROM clicks WHERE url_id = $1
		ORDER BY clicked_at DESC
		LIMIT 100`, url.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query clicks: %w", err)
	}
	defer rows.Close()

	var clicks []domain.Click
	for rows.Next() {
		var click domain.Click
		if err := rows.Scan(&click.ID, &click.UserAgent, &click.IPAddress, &click.ClickedAt); err != nil {
			return nil, fmt.Errorf("failed to scan click row: %w", err)
		}
		click.URLID = url.ID
		clicks = append(clicks, click)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clicks: %w", err)
	}

	report := &domain.AnalyticsReport{
		TotalClicks:    len(clicks),
		DailyStats:     make(map[string]int),
		MonthlyStats:   make(map[string]int),
		UserAgentStats: make(map[string]int),
		Clicks:         clicks,
	}

	for _, click := range clicks {
		date := click.ClickedAt.Format("2006-01-02")
		month := click.ClickedAt.Format("2006-01")
		report.DailyStats[date]++
		report.MonthlyStats[month]++
		report.UserAgentStats[click.UserAgent]++
	}

	return report, nil
}

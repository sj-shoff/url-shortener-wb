package usecase

import (
	"context"
	"url-shortener-wb/internal/domain"
)

type URLRepository interface {
	Create(ctx context.Context, url *domain.URL) error
	GetByAlias(ctx context.Context, alias string) (*domain.URL, error)
	ExistsByAlias(ctx context.Context, alias string) (bool, error)
}

type AnalyticsRepository interface {
	RecordClick(ctx context.Context, click *domain.Click) error
	GetAnalytics(ctx context.Context, alias string) (*domain.AnalyticsReport, error)
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Exists(ctx context.Context, key string) (bool, error)
}

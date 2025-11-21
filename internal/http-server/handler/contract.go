package handler

import (
	"context"
	"url-shortener-wb/internal/domain"
)

type URLUsecase interface {
	CreateShortURL(ctx context.Context, originalURL, customAlias string) (string, error)
	GetOriginalURL(ctx context.Context, alias string) (string, error)
	RecordClick(ctx context.Context, alias, userAgent, ip string) error
	GetAnalytics(ctx context.Context, alias string) (*domain.AnalyticsReport, error)
}

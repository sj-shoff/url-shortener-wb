package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"url-shortener-wb/internal/domain"
)

type analyticsUsecase struct {
	analyticsRepo AnalyticsRepository
	urlRepo       URLRepository
}

func NewAnalyticsUsecase(analyticsRepo AnalyticsRepository, urlRepo URLRepository) *analyticsUsecase {
	return &analyticsUsecase{
		analyticsRepo: analyticsRepo,
		urlRepo:       urlRepo,
	}
}

func (au *analyticsUsecase) RecordClick(ctx context.Context, alias, userAgent, ip string) error {
	url, err := au.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: url not found for alias %s", ErrNotFound, alias)
		}
		return fmt.Errorf("failed to get url by alias: %w", err)
	}

	click := &domain.Click{
		URLID:     url.ID,
		UserAgent: userAgent,
		IPAddress: ip,
		ClickedAt: time.Now(),
	}

	if err := au.analyticsRepo.RecordClick(ctx, click); err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	return nil
}

func (au *analyticsUsecase) GetAnalytics(ctx context.Context, alias string) (*domain.AnalyticsReport, error) {
	if alias == "" {
		return nil, fmt.Errorf("%w: empty alias", ErrInvalidAlias)
	}

	report, err := au.analyticsRepo.GetAnalytics(ctx, alias)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("%w: analytics not found for alias %s", ErrNotFound, alias)
		}
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return report, nil
}

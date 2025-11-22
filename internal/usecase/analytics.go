package usecase

import (
	"context"
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
		return err
	}

	click := &domain.Click{
		URLID:     url.ID,
		UserAgent: userAgent,
		IPAddress: ip,
		ClickedAt: time.Now(),
	}

	return au.analyticsRepo.RecordClick(ctx, click)
}

func (au *analyticsUsecase) GetAnalytics(ctx context.Context, alias string) (*domain.AnalyticsReport, error) {
	return au.analyticsRepo.GetAnalytics(ctx, alias)
}

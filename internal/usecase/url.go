package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/wb-go/wbf/zlog"

	"url-shortener-wb/internal/domain"
)

type URLUsecase struct {
	urlRepo       URLRepository
	analyticsRepo AnalyticsRepository
	cache         Cache
	logger        *zlog.Zerolog
}

func NewURLUsecase(
	urlRepo URLRepository,
	analyticsRepo AnalyticsRepository,
	cache Cache,
	logger *zlog.Zerolog,
) *URLUsecase {
	return &URLUsecase{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
		cache:         cache,
		logger:        logger,
	}
}

func (u *URLUsecase) CreateShortURL(ctx context.Context, originalURL, customAlias string) (string, error) {
	alias := customAlias
	if alias == "" {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		alias = base64.RawURLEncoding.EncodeToString(b)[:6]
	}

	exists, err := u.urlRepo.ExistsByAlias(ctx, alias)
	if err != nil {
		return "", err
	}

	if exists {
		return "", errors.New("alias already exists")
	}

	url := &domain.URL{
		OriginalURL: originalURL,
		Alias:       alias,
		CreatedAt:   time.Now(),
	}

	if err := u.urlRepo.Create(ctx, url); err != nil {
		return "", err
	}

	if err := u.cache.Set(ctx, alias, originalURL); err != nil {
		u.logger.Warn().Err(err).Str("alias", alias).Msg("failed to cache URL")
	}

	return alias, nil
}

func (u *URLUsecase) GetOriginalURL(ctx context.Context, alias string) (string, error) {
	originalURL, err := u.cache.Get(ctx, alias)
	if err == nil {
		u.logger.Debug().Str("alias", alias).Msg("cache hit")
		return originalURL, nil
	}

	url, err := u.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		return "", err
	}

	if err := u.cache.Set(ctx, alias, url.OriginalURL); err != nil {
		u.logger.Warn().Err(err).Str("alias", alias).Msg("failed to update cache")
	}

	return url.OriginalURL, nil
}

func (u *URLUsecase) RecordClick(ctx context.Context, alias, userAgent, ip string) error {
	url, err := u.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		return err
	}

	click := &domain.Click{
		URLID:     url.ID,
		UserAgent: userAgent,
		IPAddress: ip,
		ClickedAt: time.Now(),
	}

	return u.analyticsRepo.RecordClick(ctx, click)
}

func (u *URLUsecase) GetAnalytics(ctx context.Context, alias string) (*domain.AnalyticsReport, error) {
	return u.analyticsRepo.GetAnalytics(ctx, alias)
}

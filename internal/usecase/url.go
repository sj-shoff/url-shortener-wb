package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"url-shortener-wb/internal/domain"

	"github.com/wb-go/wbf/zlog"
)

var (
	customAliasRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`)
)

type urlUsecase struct {
	urlRepo URLRepository
	cache   Cache
	logger  *zlog.Zerolog
}

func NewURLUsecase(
	urlRepo URLRepository,
	cache Cache,
	logger *zlog.Zerolog,
) *urlUsecase {
	return &urlUsecase{
		urlRepo: urlRepo,
		cache:   cache,
		logger:  logger,
	}
}

func validateURL(originalURL string) error {
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	return nil
}

func validateCustomAlias(alias string) error {
	if alias == "" {
		return nil
	}
	if !customAliasRegex.MatchString(alias) {
		return fmt.Errorf("%w: alias must be 3-20 alphanumeric characters", ErrInvalidAlias)
	}
	return nil
}

func (u *urlUsecase) CreateShortURL(ctx context.Context, originalURL, customAlias string) (string, error) {
	if err := validateURL(originalURL); err != nil {
		return "", err
	}

	if err := validateCustomAlias(customAlias); err != nil {
		return "", err
	}

	alias := customAlias
	if alias == "" {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("failed to generate random alias: %w", err)
		}
		alias = base64.RawURLEncoding.EncodeToString(b)[:6]
	}

	exists, err := u.urlRepo.ExistsByAlias(ctx, alias)
	if err != nil {
		return "", fmt.Errorf("failed to check alias existence: %w", err)
	}

	if exists {
		return "", fmt.Errorf("%w: alias %s", ErrAliasExists, alias)
	}

	url := &domain.URL{
		OriginalURL: originalURL,
		Alias:       alias,
		CreatedAt:   time.Now(),
	}

	if err := u.urlRepo.Create(ctx, url); err != nil {
		return "", fmt.Errorf("failed to create url: %w", err)
	}

	if err := u.cache.Set(ctx, alias, originalURL); err != nil {
		u.logger.Warn().Err(err).Str("alias", alias).Msg("failed to cache URL")
	}

	return alias, nil
}

func (u *urlUsecase) GetOriginalURL(ctx context.Context, alias string) (string, error) {
	if alias == "" {
		return "", fmt.Errorf("%w: empty alias", ErrInvalidAlias)
	}

	originalURL, err := u.cache.Get(ctx, alias)
	if err == nil {
		u.logger.Debug().Str("alias", alias).Msg("cache hit")
		return originalURL, nil
	}

	url, err := u.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", fmt.Errorf("%w: url not found for alias %s", ErrNotFound, alias)
		}
		return "", fmt.Errorf("failed to get url by alias: %w", err)
	}

	if err := u.cache.Set(ctx, alias, url.OriginalURL); err != nil {
		u.logger.Warn().Err(err).Str("alias", alias).Msg("failed to update cache")
	}

	return url.OriginalURL, nil
}

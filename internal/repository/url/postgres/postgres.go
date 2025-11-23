package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"url-shortener-wb/internal/domain"
	repo "url-shortener-wb/internal/repository"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type URLRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewURLRepository(
	db *dbpg.DB,
	retries retry.Strategy,
) *URLRepository {
	return &URLRepository{
		db:      db,
		retries: retries,
	}
}

func (r *URLRepository) Create(ctx context.Context, url *domain.URL) error {
	_, err := r.db.ExecWithRetry(ctx, r.retries,
		`INSERT INTO urls (original_url, alias, created_at)
		VALUES ($1, $2, $3) RETURNING id`,
		url.OriginalURL, url.Alias, url.CreatedAt,
	)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"urls_alias_key\"" {
			return fmt.Errorf("%w: alias %s already exists", repo.ErrAlreadyExists, url.Alias)
		}
		return fmt.Errorf("failed to insert url: %w", err)
	}
	return nil
}

func (r *URLRepository) GetByAlias(ctx context.Context, alias string) (*domain.URL, error) {
	row, err := r.db.QueryRowWithRetry(ctx, r.retries,
		`SELECT id, original_url, alias, created_at
		FROM urls WHERE alias = $1 LIMIT 1`, alias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: alias %s not found", repo.ErrNotFound, alias)
		}
		return nil, fmt.Errorf("failed to query url by alias: %w", err)
	}

	var url domain.URL
	if err := row.Scan(&url.ID, &url.OriginalURL, &url.Alias, &url.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: alias %s not found", repo.ErrNotFound, alias)
		}
		return nil, fmt.Errorf("failed to scan url row: %w", err)
	}

	return &url, nil
}

func (r *URLRepository) ExistsByAlias(ctx context.Context, alias string) (bool, error) {
	var exists bool
	err := retry.DoContext(ctx, r.retries, func() error {
		row := r.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM urls WHERE alias = $1)`, alias)

		if err := row.Scan(&exists); err != nil {
			return fmt.Errorf("failed to scan exists query: %w", err)
		}
		return nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to check alias existence: %w", err)
	}
	return exists, nil
}

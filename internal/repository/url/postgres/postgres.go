package postgres

import (
	"context"
	"database/sql"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"

	"url-shortener-wb/internal/domain"
	repo "url-shortener-wb/internal/repository"
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
	return err
}

func (r *URLRepository) GetByAlias(ctx context.Context, alias string) (*domain.URL, error) {
	row, err := r.db.QueryRowWithRetry(ctx, r.retries,
		`SELECT id, original_url, alias, created_at
		FROM urls WHERE alias = $1 LIMIT 1`, alias)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}

	var url domain.URL
	if err := row.Scan(&url.ID, &url.OriginalURL, &url.Alias, &url.CreatedAt); err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLRepository) ExistsByAlias(ctx context.Context, alias string) (bool, error) {
	var exists bool
	retryErr := retry.Do(func() error {
		row := r.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM urls WHERE alias = $1)`, alias)

		if scanErr := row.Scan(&exists); scanErr != nil {
			return scanErr
		}
		return nil
	}, r.retries)

	if retryErr != nil {
		return false, retryErr
	}
	return exists, nil
}

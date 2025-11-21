package domain

import "time"

type URL struct {
	ID          int64
	OriginalURL string
	Alias       string
	CreatedAt   time.Time
}

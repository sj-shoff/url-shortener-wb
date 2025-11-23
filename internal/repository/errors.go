package repository

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrCacheMiss     = errors.New("cache miss")
	ErrDBConnection  = errors.New("database connection failed")
)

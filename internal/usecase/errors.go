package usecase

import "errors"

var (
	ErrNotFound     = errors.New("url not found")
	ErrAliasExists  = errors.New("alias already exists")
	ErrInvalidURL   = errors.New("invalid url format")
	ErrInvalidAlias = errors.New("invalid alias format")
)

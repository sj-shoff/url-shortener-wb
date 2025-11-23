package handler

import "errors"

var (
	ErrAliasExists  = errors.New("alias already exists")
	ErrInvalidAlias = errors.New("invalid alias parameter")
	ErrInvalidURL   = errors.New("invalid URL format")
	ErrNotFound     = errors.New("resource not found")
)

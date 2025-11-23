package handler

import "errors"

var (
	ErrInvalidRequest     = errors.New("invalid request body")
	ErrAliasExists        = errors.New("alias already exists")
	ErrInvalidAlias       = errors.New("invalid alias parameter")
	ErrInvalidURL         = errors.New("invalid URL format")
	ErrInternalServer     = errors.New("internal server error")
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidCustomAlias = errors.New("invalid custom alias format")
)

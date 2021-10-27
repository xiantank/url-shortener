package errors

import "errors"

var (
	ErrExpired  = errors.New("expired")
	ErrNotFound = errors.New("not found")
)

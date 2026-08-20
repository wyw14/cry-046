package service

import "errors"

var (
	ErrInvalid   = errors.New("invalid request")
	ErrForbidden = errors.New("operation forbidden")
	ErrNotFound  = errors.New("resource not found")
	ErrExpired   = errors.New("resource expired")
)

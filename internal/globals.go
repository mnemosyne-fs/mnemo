package internal

import "errors"

var (
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotExists  = errors.New("user does not exist")
	ErrInvalidLogin   = errors.New("invalid login")
	ErrInvalidSession = errors.New("invalid session")
)

const (
	FilePerm         = 0644
	SESSION_LIFETIME = 604800
)

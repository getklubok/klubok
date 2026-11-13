package domain

import (
	"errors"
	"time"
)

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
)

type Session struct {
	ID        SessionID `json:"id"`
	CookieID  CookieID  `json:"-"`
	UserID    UserID    `json:"-"`
	CheckedAt time.Time `json:"checkedAt"`
	IsCurrent bool      `json:"isCurrent"`
}

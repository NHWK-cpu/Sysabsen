package models

import "time"

type UserDevice struct {
	ID                uint64    `json:"id"`
	UserID            uint64    `json:"user_id"`
	DeviceCookieToken string    `json:"device_cookie_token"`
	UserAgent         string    `json:"user_agent"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}
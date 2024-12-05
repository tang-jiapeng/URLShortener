package model

import "time"

type CreateURLRequest struct {
	OriginalURL string `json:"original_url"`
	CustomCode  string `json:"custom_code"`
	Duration    *int   `json:"duration"`
}

type CreateURLResponse struct {
	ShortURL  string    `json:"short_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

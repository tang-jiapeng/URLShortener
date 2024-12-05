// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package repository

import (
	"context"
)

type Querier interface {
	CreateURL(ctx context.Context, arg CreateURLParams) error
	DeleteExpiredURLs(ctx context.Context) error
	GetCreatedURL(ctx context.Context) (Url, error)
	GetURLByShortCode(ctx context.Context, shortCode string) (Url, error)
	IsShortCodeAvailable(ctx context.Context, shortCode string) (bool, error)
}

var _ Querier = (*Queries)(nil)
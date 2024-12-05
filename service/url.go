package service

import (
	"URLShortener/cache"
	"URLShortener/config"
	"URLShortener/model"
	"URLShortener/pkg/shortcode"
	"URLShortener/repository"
	"context"
	"database/sql"
	"errors"
	"time"
)

type URLService interface {
	CreateURL(ctx context.Context, params model.CreateURLRequest) (*repository.Url, error)
	GetURL(ctx context.Context, shortCode string) (*repository.Url, error)
	Cleanup(ctx context.Context) error
}

type urlService struct {
	querier   repository.Querier
	cache     cache.Cache
	generator shortcode.Generator
	db        *sql.DB
	cfg       *config.Config
}

func (u *urlService) tryFiveIsAvaliable(ctx context.Context, n int) (string, error) {
	if n >= 5 {
		return "", errors.New("try 5 times and failed")
	}
	shortCode := u.generator.NextID()
	isAvaliable, err := u.querier.IsShortCodeAvailable(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if !isAvaliable {
		return u.tryFiveIsAvaliable(ctx, n+1)
	}
	return shortCode, nil
}

func (u *urlService) CreateURL(ctx context.Context, params model.CreateURLRequest) (*repository.Url, error) {
	var isCustom bool
	var shortCode string
	var expiresAt time.Time
	if params.CustomCode != "" {
		isAvaliable, err := u.querier.IsShortCodeAvailable(ctx, params.CustomCode)
		if err != nil {
			return nil, err
		}
		if !isAvaliable {
			return nil, errors.New("custom code is not available")
		}
		shortCode = params.CustomCode
		isCustom = true
	} else {
		var err error
		shortCode, err = u.tryFiveIsAvaliable(ctx, 0)
		if err != nil {
			return nil, err
		}
	}
	if params.Duration == nil {
		expiresAt = time.Now().Add(u.cfg.App.DefaultExpiration)
	} else {
		expiresAt = time.Now().Add(time.Hour * time.Duration(*params.Duration))
	}
	err := u.querier.CreateURL(ctx, repository.CreateURLParams{
		OriginalUrl: params.OriginalURL,
		ShortCode:   shortCode,
		ExpiresAt:   expiresAt,
		IsCustom:    isCustom,
	})
	if err != nil {
		return nil, err
	}
	url, err := u.querier.GetCreatedURL(ctx)
	if err != nil {
		return nil, err
	}
	if err = u.cache.SetURL(ctx, url); err != nil {
		return nil, err
	}
	return &url, nil
}

func (u *urlService) GetURL(ctx context.Context, shortCode string) (*repository.Url, error) {
	url, err := u.cache.GetURL(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if url != nil {
		return url, nil
	}
	url2, err := u.querier.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if err := u.cache.SetURL(ctx, url2); err != nil {
		return nil, err
	}
	return &url2, nil
}

func (u *urlService) Cleanup(ctx context.Context) error {
	return u.querier.DeleteExpiredURLs(ctx)
}

func NewUrlService(db *sql.DB, cache cache.Cache, generator shortcode.Generator, cfg *config.Config) *urlService {
	return &urlService{
		querier:   repository.New(db),
		cache:     cache,
		generator: generator,
		db:        db,
		cfg:       cfg,
	}
}

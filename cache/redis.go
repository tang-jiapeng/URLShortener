package cache

import (
	"URLShortener/config"
	"URLShortener/repository"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

const urlKeyPrefix = "url:"

type Cache interface {
	GetURL(ctx context.Context, shortCode string) (*repository.Url, error)
	SetURL(ctx context.Context, url repository.Url) error
	DeleteURL(ctx context.Context, shortCode string) error
	Close() error
}

type redisCache struct {
	client *redis.Client
}

func (r *redisCache) GetURL(ctx context.Context, shortCode string) (*repository.Url, error) {
	key := urlKeyPrefix + shortCode
	data, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var url repository.Url
	if err = json.Unmarshal(data, &url); err != nil {
		return nil, err
	}
	if url.ExpiresAt.Before(time.Now()) {
		r.client.Del(ctx, key)
		return nil, nil
	}
	return &url, nil
}

func (r *redisCache) SetURL(ctx context.Context, url repository.Url) error {
	key := urlKeyPrefix + url.ShortCode
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	if url.ExpiresAt.Before(time.Now()) {
		return nil
	}
	return r.client.Set(ctx, key, data, time.Until(url.ExpiresAt)).Err()
}

func (r *redisCache) DeleteURL(ctx context.Context, shortCode string) error {
	key := urlKeyPrefix + shortCode
	return r.client.Del(ctx, key).Err()
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

func NewRedisCache(cfg config.RedisConfig) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &redisCache{client: client}, nil
}

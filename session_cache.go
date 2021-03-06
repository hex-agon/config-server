package main

import (
	"context"
	"github.com/dgraph-io/ristretto"
	"time"
)

type SessionCache interface {
	GetUserId(context context.Context, uuid string) (int64, error)
}

type RistSessionCache struct {
	repository SessionRepository
	cache      *ristretto.Cache
}

func NewSessionCache(repository SessionRepository, cacheSize int64) (SessionCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		// More info at https://github.com/dgraph-io/ristretto#config
		NumCounters: cacheSize * 10,
		MaxCost:     cacheSize,
		BufferItems: 64,
		OnEvict: func(item *ristretto.Item) {
			go updateEvicted(item.Value.(int64), repository)
		},
	})

	if err != nil {
		return nil, err
	}
	return &RistSessionCache{
		repository: repository,
		cache:      cache,
	}, nil
}

func (c *RistSessionCache) GetUserId(ctx context.Context, uuid string) (int64, error) {
	value, hit := c.cache.Get(uuid)
	if hit {
		return value.(int64), nil
	}
	userId, err := c.repository.FindUserIdByUuid(ctx, uuid)
	if err != nil {
		return -1, err
	}
	c.cache.SetWithTTL(uuid, userId, 1, 30*time.Minute)
	return userId, nil
}

func updateEvicted(userId int64, repository SessionRepository) {
	_ = repository.UpdateLastUsedByUserId(userId)
}

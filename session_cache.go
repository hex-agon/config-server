package main

import "github.com/dgraph-io/ristretto"

type SessionCache interface {
	GetUserId(uuid string) (int64, error)
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
			go updateEvicted(item.Value.(string), repository)
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

func (c *RistSessionCache) GetUserId(uuid string) (int64, error) {
	value, hit := c.cache.Get(uuid)
	if hit {
		return value.(int64), nil
	}
	userId, err := c.repository.FindUserIdByUuid(uuid)
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func updateEvicted(uuid string, repository SessionRepository) {
	_ = repository.UpdateLastUsedByUuid(uuid)
}

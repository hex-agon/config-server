package main

import "context"

type ConfigRepository interface {
	FindByUserId(ctx context.Context, userId int64) (*Configuration, error)
	Save(ctx context.Context, userId int64, entry *ConfigEntry) error
	SaveBatch(ctx context.Context, userId int64, configuration *Configuration) ([]string, error)
	DeleteKey(ctx context.Context, userId int64, key string) error
}

type Configuration struct {
	Config []ConfigEntry `json:"config"`
}

type ConfigEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SessionRepository interface {
	FindUserIdByUuid(ctx context.Context, uuid string) (int64, error)
	UpdateLastUsedByUuid(uuid string) error
}

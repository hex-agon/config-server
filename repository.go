package main

type ConfigRepository interface {
	FindByUserId(userId int64) (*Configuration, error)
	Save(userId int64, entry *ConfigEntry) error
	SaveBatch(userId int64, configuration *Configuration) ([]string, error)
	DeleteKey(userId int64, key string) error
}

type Configuration struct {
	Config []ConfigEntry `json:"config"`
}

type ConfigEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SessionRepository interface {
	FindUserIdByUuid(uuid string) (int64, error)
}

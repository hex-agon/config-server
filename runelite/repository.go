package runelite

type ConfigRepository interface {
	FindByUserId(userId int) (*Configuration, error)
	Save(userId int, entry *ConfigEntry) error
	SaveBatch(userId int, configuration *Configuration) error
	DeleteKey(userId int, key string) error
}

type Configuration struct {
	Config []ConfigEntry `json:"config"`
}

type ConfigEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

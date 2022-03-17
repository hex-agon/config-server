package runelite

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

type mongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(collection *mongo.Collection) ConfigRepository {
	return &mongoRepository{
		collection: collection,
	}
}

func invalidConfigKey(key string) bool {
	// $ and _ are reserved prefixes for mongodb
	return key == "" || strings.HasPrefix(key, "$") || strings.HasPrefix(key, "_")
}

func sanitizeConfigKey(key string) string {
	return strings.ReplaceAll(key, ".", ":")
}

func serializeGroup(groupKey string, group interface{}) []ConfigEntry {
	groupMap := group.(map[string]interface{})
	entries := make([]ConfigEntry, len(groupMap))
	for key, value := range groupMap {
		valueJson, err := serializeGroupValue(value)

		if err != nil {
			continue
		}
		entries = append(entries, ConfigEntry{
			Key:   groupKey + "." + sanitizeConfigKey(key),
			Value: valueJson,
		})
	}
	return entries
}

func serializeGroupValue(value interface{}) (string, error) {
	switch value.(type) {
	case string:
		return value.(string), nil
	default:
		marshal, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(marshal), nil
	}
}

func (m *mongoRepository) FindByUserId(userId int) (*Configuration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var document map[string]interface{}

	err := m.collection.FindOne(
		ctx,
		bson.M{"_userId": userId},
		options.FindOne().SetProjection(bson.M{"_id": 0, "_userId": 0}),
	).Decode(&document)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		entries := make([]ConfigEntry, 0)
		for groupKey, group := range document {
			entries = append(entries, serializeGroup(groupKey, group)...)
		}
		configuration := &Configuration{
			Config: entries,
		}
		return configuration, nil
	}
}

func (m *mongoRepository) Save(userId int, entry *ConfigEntry) error {
	key := entry.Key

	if invalidConfigKey(key) {
		return errors.New("invalid config key")
	}
	update := bson.M{"$set": bson.M{sanitizeConfigKey(key): entry.Value}}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := m.collection.UpdateOne(ctx, bson.M{"_userId": userId}, update, options.Update().SetUpsert(true))

	return err
}

func (m *mongoRepository) SaveBatch(userId int, configuration *Configuration) error {
	entries := bson.M{}
	// filter invalid keys
	for _, entry := range configuration.Config {
		if invalidConfigKey(entry.Key) {
			continue
		}
		entries[sanitizeConfigKey(entry.Key)] = entry.Value
	}
	update := bson.M{"$set": entries}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.collection.UpdateOne(ctx, bson.M{"_userId": userId}, update, options.Update().SetUpsert(true))
	return err
}

func (m *mongoRepository) DeleteKey(userId int, key string) error {
	if invalidConfigKey(key) {
		return errors.New("invalid config key")
	}
	unset := bson.M{"$unset": bson.M{sanitizeConfigKey(key): nil}}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.collection.UpdateOne(ctx, bson.M{"_userId": userId}, unset)
	return err
}

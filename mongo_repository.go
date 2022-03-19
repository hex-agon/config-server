package main

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strings"
	"time"
)

var maybeJsonPattern = regexp.MustCompile("^[\\-0-9{\\[\"]|true|false")

type mongoConfigRepository struct {
	collection           *mongo.Collection
	maxConfigValueLength int64
}

func NewConfigRepository(collection *mongo.Collection, maxConfigValueLength int64) ConfigRepository {
	return &mongoConfigRepository{
		collection:           collection,
		maxConfigValueLength: maxConfigValueLength,
	}
}

func invalidConfigKey(key string) bool {
	// $ and _ are reserved prefixes for mongodb
	return key == "" || strings.HasPrefix(key, "$") || strings.HasPrefix(key, "_")
}

func sanitizeConfigKey(key string) string {
	parts := strings.SplitN(key, ".", 2)

	if len(parts) != 2 {
		return ""
	}
	return parts[0] + "." + strings.ReplaceAll(parts[1], ".", ":")
}

func serializeGroup(groupKey string, group interface{}) []ConfigEntry {
	groupMap := group.(map[string]interface{})
	entries := make([]ConfigEntry, len(groupMap))
	idx := 0
	for key, value := range groupMap {
		serializedValue, err := serializeGroupValue(value)

		if err != nil {
			continue
		}
		entries[idx].Key = groupKey + "." + key
		entries[idx].Value = serializedValue
		idx++
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

func deserializeGroupValue(value string, maxLength int64) (interface{}, error) {

	if int64(len(value)) > maxLength {
		return nil, errors.New("value exceeds max length")
	} else if maybeJsonPattern.MatchString(value) {
		// this should be safe to deserialize without much validation https://go-review.googlesource.com/c/go/+/199837
		var deserializedValue interface{}
		err := json.Unmarshal([]byte(value), &deserializedValue)
		if err != nil {
			return nil, err
		}
		return deserializedValue, nil
	}
	return value, nil
}

func (m *mongoConfigRepository) FindByUserId(userId int64) (*Configuration, error) {
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

func (m *mongoConfigRepository) Save(userId int64, entry *ConfigEntry) error {
	key := entry.Key

	if invalidConfigKey(key) {
		return errors.New("invalid config key")
	}
	value, err := deserializeGroupValue(entry.Value, m.maxConfigValueLength)

	if err != nil {
		return err
	}
	update := bson.M{"$set": bson.M{sanitizeConfigKey(key): value}}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = m.collection.UpdateOne(ctx, bson.M{"_userId": userId}, update, options.Update().SetUpsert(true))

	return err
}

func (m *mongoConfigRepository) SaveBatch(userId int64, configuration *Configuration) ([]string, error) {
	entries := bson.M{}
	failedKeys := make([]string, 0)
	for _, entry := range configuration.Config {
		if invalidConfigKey(entry.Key) {
			failedKeys = append(failedKeys, entry.Key)
			continue
		}
		value, err := deserializeGroupValue(entry.Value, m.maxConfigValueLength)

		if err != nil {
			failedKeys = append(failedKeys, entry.Key)
		} else {
			entries[sanitizeConfigKey(entry.Key)] = value
		}
	}
	update := bson.M{"$set": entries}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.collection.UpdateOne(ctx, bson.M{"_userId": userId}, update, options.Update().SetUpsert(true))
	return failedKeys, err
}

func (m *mongoConfigRepository) DeleteKey(userId int64, key string) error {
	if invalidConfigKey(key) {
		return errors.New("invalid config key")
	}
	unset := bson.M{"$unset": bson.M{sanitizeConfigKey(key): nil}}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.collection.UpdateOne(ctx, bson.M{"_userId": userId}, unset)
	return err
}

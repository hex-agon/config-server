package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestSerializeGroupValue(t *testing.T) {
	tests := []struct {
		serializable interface{}
		expected     string
	}{
		{serializable: "MITCHELL", expected: "MITCHELL"},
		{serializable: 1251, expected: "1251"},
		{serializable: -16777216, expected: "-16777216"},
		{
			serializable: []string{"NOON", "PET_SMOKE_DEVIL", "VORKI"},
			expected:     "[\"NOON\",\"PET_SMOKE_DEVIL\",\"VORKI\"]",
		},
		{
			serializable: struct {
				Type  string `json:"type"`
				Name  string `json:"name"`
				Kills int    `json:"kills"`
			}{
				Type:  "EVENT",
				Name:  "Chambers of Xeric",
				Kills: 4,
			},
			expected: "{\"type\":\"EVENT\",\"name\":\"Chambers of Xeric\",\"kills\":4}",
		},
	}

	for _, test := range tests {
		if value, err := serializeGroupValue(test.serializable); err != nil || value != test.expected {
			t.Errorf("Got serialized value %s but expected %s", value, test.expected)
		}
	}
}

func TestDeserializeGroupValue(t *testing.T) {
	tests := []struct {
		serialized string
		expected   interface{}
	}{
		{serialized: "string", expected: "string"},
		{serialized: "string with spaces", expected: "string with spaces"},
		{serialized: "true", expected: true},
		{serialized: "false", expected: false},
		{serialized: "1", expected: float64(1)},
		{serialized: "1.2", expected: 1.2},
		{serialized: "\"quote\"", expected: "quote"},
		{serialized: "{\"key\": \"value\"}", expected: map[string]interface{}{"key": "value"}},
	}
	for _, test := range tests {

		switch test.expected.(type) {
		case map[string]interface{}:
			if value, err := deserializeGroupValue(test.serialized, 1024); err != nil || !reflect.DeepEqual(value, test.expected) {
				t.Errorf("Got deserialized %s but expected %s", value, test.expected)
			}
		default:
			if value, err := deserializeGroupValue(test.serialized, 1024); err != nil || value != test.expected {
				t.Errorf("Got deserialized value %s but expected %s", value, test.expected)
			}
		}
	}
}

func TestDeserializeGroupValueExceedMaxLength(t *testing.T) {
	jsonBomb := strings.Repeat("{\"a\":", 1024) + "[]" + strings.Repeat("}", 1024)

	if value, err := deserializeGroupValue(jsonBomb, 128); err == nil || value != nil {
		t.Errorf("Deserialized json bomb string")
	}
}

func TestMaybeJson(t *testing.T) {
	tests := []struct {
		value string
		json  bool
	}{
		{value: "string", json: false},
		{value: "string with spaces", json: false},
		{value: "true", json: true},
		{value: "false", json: true},
		{value: "1", json: true},
		{value: "1.2", json: true},
		{value: "\"quote\"", json: true},
		{value: "{\"key\": \"value\"}", json: true},
		{value: "[42]", json: true},
	}
	for _, test := range tests {
		if isJson := maybeJsonPattern.MatchString(test.value); isJson != test.json {
			t.Errorf("Matched value as %t but expected %t", isJson, test.json)
		}
	}
}

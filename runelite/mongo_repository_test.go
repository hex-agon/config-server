package runelite

import (
	"testing"
)

func TestSerializeGroupValue(t *testing.T) {
	tests := []struct {
		serializable interface{}
		expected     string
	}{
		{
			serializable: "MITCHELL",
			expected:     "MITCHELL",
		},
		{
			serializable: 1251,
			expected:     "1251",
		},
		{
			serializable: -16777216,
			expected:     "-16777216",
		},
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

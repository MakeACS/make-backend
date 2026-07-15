package acsmqtt

import (
	"testing"
)

func TestKnownTopicTableDriven(t *testing.T) {
	// Defining the columns of the table

	var tests = []struct {
		name           string
		requestedTopic string
		id             string
		want           bool
	}{
		// the table itself
		{"Matching ID in correct place", "a/b/12/d", "12", true},
		{"Matching ID in unknown topic", "a/b/12/e", "12", false},
		{"Non-matching ID in correct place", "a/b/12/d", "13", false},
		{"Non-matching ID in unknown topic", "a/b/12/e", "13", false},
	}

	var knownTopics []SNPair = []SNPair{
		{"a/b/+/f", 6}, // invalid topic
		{"a/b/+/d", 2}, // valid topic
	}

	// Check All
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := doesTopicMatchAllowed(tt.id, tt.requestedTopic, knownTopics)
			if ans != tt.want {
				t.Errorf("test '%s': got %v, want %v", tt.name, ans, tt.want)
			}
		})
	}
}

func TestValidDevicePubTopics(t *testing.T) {
	// Defining the columns of the table

	// Check All
	for _, pair := range knownDevicePubTopics {
		t.Run(pair.wildcardTopic, func(t *testing.T) {
			if !pair.isValid() {
				t.Errorf("pub topic matcher for %s has sn location %d that exceeds its number of parts", pair.wildcardTopic, pair.snLocation)
			}
		})
	}
}
func TestValidDeviceSubTopics(t *testing.T) {
	// Defining the columns of the table

	// Check All
	for _, pair := range knownDeviceSubTopics {
		t.Run(pair.wildcardTopic, func(t *testing.T) {
			if !pair.isValid() {
				t.Errorf("pub topic matcher for %s has sn location %d that exceeds its number of parts", pair.wildcardTopic, pair.snLocation)
			}
		})
	}
}

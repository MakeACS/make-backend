package acsmqtt

import (
	"bytes"
	"log/slog"
	"strings"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type AuthHook struct {
	mqtt.HookBase
}

// ID returns the ID of the hook.
func (h *AuthHook) ID() string {
	return "acs-auth"
}

// Provides indicates which hook methods this hook provides.
func (h *AuthHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnACLCheck,
	}, []byte{b})
}

// OnConnectAuthenticate returns true/allowed for all requests.
func (h *AuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	return true
}

// OnACLCheck returns true/allowed for all checks.
func (h *AuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	if cl.ID == serverUsername {
		// allow server access to all channels
		return true
	}
	// otherwise, assume you're a device
	// only allow if it matches a channel we know and your SN aligns with
	if write {
		return isDeviceAllowedToPublish(cl.ID, topic)
	} else {
		return isDeviceAllowedToSubscribe(cl.ID, topic)
	}
}

type SNPair struct {
	wildcardTopic string
	snLocation    int
}

// returns true if the request topic parts matches the wildcarded topic
// if so, return the part of the requested topic that needs to be checked against the ID of the cliend
func (pair SNPair) matchesTopic(requestedParts []string) (bool, string) {
	var knownParts = strings.Split(pair.wildcardTopic, "/")

	// ignore invalid
	if pair.snLocation > len(knownParts)-1 {
		slog.Warn("Invalid topic matcher for mqtt auth", "topic", pair.wildcardTopic, "location", pair.snLocation)
		return false, ""
	}
	if pair.snLocation > len(requestedParts)-1 {
		return false, ""
	}
	for i, part := range requestedParts {
		if i == pair.snLocation {
			// wildcard will not match real value, need to return that to check
			continue
		}
		if part != knownParts[i] {
			// doesn't even apply
			return false, ""
		}
	}
	return true, requestedParts[pair.snLocation]
}

func doesTopicMatchAllowed(ID string, topic string, allowed []SNPair) bool {
	var requestedParts = strings.Split(topic, "/")
	for _, pair := range allowed {
		applies, toCheck := pair.matchesTopic(requestedParts)
		if !applies {
			continue
		}
		if toCheck == ID {
			return true
		} else {
			return false
		}

	}
	// is not a known topic
	return false
}
func isDeviceAllowedToPublish(ID string, topic string) bool {
	var knownDevicePubTopics []SNPair = []SNPair{
		SNPair{"makerspace/device/+/status", 2},
		SNPair{"makerspace/device/+/stateChange", 2},
		SNPair{"makerspace/device/+/log", 2},
		SNPair{"makerspace/device/+/authTo/request", 2},
		SNPair{"makerspace/device/+/config/report", 2},
		SNPair{"makerspace/device/+/info/request", 2},
		SNPair{"makerspace/device/+/welcome/request", 2},
	}
	return doesTopicMatchAllowed(ID, topic, knownDevicePubTopics)
}
func isDeviceAllowedToSubscribe(ID string, topic string) bool {
	var knownDeviceSubTopics []SNPair = []SNPair{
		SNPair{"makerspace/device/+/status", 2},
		SNPair{"makerspace/device/+/stateChange", 2},
		SNPair{"makerspace/device/+/log", 2},
		SNPair{"makerspace/device/+/authTo/request", 2},
		SNPair{"makerspace/device/+/config/report", 2},
		SNPair{"makerspace/device/+/info/request", 2},
		SNPair{"makerspace/device/+/welcome/request", 2},
	}
	return doesTopicMatchAllowed(ID, topic, knownDeviceSubTopics)
}

package acsmqtt

import (
	"bytes"
	"context"
	"log/slog"
	"make-backend/internal/database"
	"strings"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

var knownDevicePubTopics []SNPair = []SNPair{
	{"makerspace/device/+/status", 2},
	{"makerspace/device/+/stateChange", 2},
	{"makerspace/device/+/log", 2},
	{"makerspace/device/+/authTo/request", 2},
	{"makerspace/device/+/config/report", 2},
	{"makerspace/device/+/info/request", 2},
	{"makerspace/device/+/welcome/request", 2},
}

var knownDeviceSubTopics []SNPair = []SNPair{
	// topics server gives
	{"makerspace/device/+/welcome/response", 2},
	{"makerspace/device/+/command", 2},
	{"makerspace/device/+/info/response", 2},
	{"makerspace/device/+/config/update", 2},
	{"makerspace/device/+/authTo/response", 2},
}

type AuthHook struct {
	mqtt.HookBase
	store *database.Store
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

// OnConnectAuthenticate returns true/allowed if server or a registered device.
func (h *AuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	if string(cl.Properties.Username) == serverUsername {
		return string(pk.Connect.Password) == serverPassword
	}

	sn := string(cl.Properties.Username)
	dev, err := h.store.Devices.GetDeviceBySN(context.TODO(), sn)
	if err != nil {
		slog.Warn("failed to authenticate device", "sn", sn, "err", err)
		return false
	}
	ok, err := dev.CredentialsMatch(string(pk.Connect.Password))
	if err != nil {
		slog.Warn("failed to check credentials on device", "sn", sn, "err", err)
		return false
	}
	return ok
}

// OnACLCheck returns true/allowed for all checks.
func (h *AuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	username := string(cl.Properties.Username)
	if username == serverUsername {
		// allow server access to all channels
		return true
	}
	// otherwise, assume you're a device
	// only allow if it matches a channel we know and your SN aligns with
	var ok bool
	if write {
		ok = doesTopicMatchAllowed(username, topic, knownDevicePubTopics)
	} else {
		ok = doesTopicMatchAllowed(username, topic, knownDeviceSubTopics)
	}

	return ok
}

type SNPair struct {
	wildcardTopic string
	snLocation    int
}

func (pair SNPair) isValid() bool {
	return pair.snLocation < len(strings.Split(pair.wildcardTopic, "/"))
}

// returns true if the request topic parts matches the wildcard topic
// if so, return the part of the requested topic that needs to be checked against the ID of the client
func (pair SNPair) matchesTopic(requestedParts []string) (bool, string) {
	var knownParts = strings.Split(pair.wildcardTopic, "/")

	// ignore invalid
	if !pair.isValid() {
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

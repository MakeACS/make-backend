package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	Id           string
	Timestamp    time.Time
	MakerspaceID *int

	PlainString  string
	FormatString string

	MessageType string
	Data        json.RawMessage
}

type LogEntity struct {
	Id    int
	Label string
}

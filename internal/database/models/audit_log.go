package models

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	Id           int
	Timestamp    time.Time
	MakerspaceID *int

	PlainString  string
	FormatString string

	MessageType string
	Data        *json.RawMessage
}

type LogEntity struct {
	Id    int
	Label string
}

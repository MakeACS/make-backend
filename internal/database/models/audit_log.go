package models

import (
	"time"
)

type AuditLog struct {
	Id           string
	Timestamp    time.Time
	MakerspaceID *int

	PlainString  string
	FormatString string

	MessageType string
	Data        map[string]any
}

type LogEntity struct {
	Id    int
	Label string
}

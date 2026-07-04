package models

import (
	"database/sql"
	"time"
)

type User struct {
	Id            int
	Username      string
	Firstname     string
	Lastname      string
	Pronouns      string
	JoinDate      time.Time
	SetupComplete bool
	Archived      bool
	Notes         string
	Admin         bool
	ForceArchive  sql.NullBool
	CardTag       string
}

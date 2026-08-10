package models

import (
	"time"
)

// User refers to a single person from the backends point of view
// It intentionally does not refer to the sign in method (SSO, OAUTH, etc) as a User is independent of the sign in method
type User struct {
	Id            int
	Email         string
	Firstname     string
	Lastname      string
	Pronouns      string
	JoinDate      time.Time
	SetupComplete bool
	Archived      bool
	Notes         string
	Admin         bool
	ForceArchive  *bool
	CardTag       string
}

func (u User) FullName() string {
	return u.Firstname + " " + u.Lastname
}
func (u User) LogEntity() LogEntity {
	return LogEntity{Id: u.Id, Label: u.FullName()}
}

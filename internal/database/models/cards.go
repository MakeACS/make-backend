package models

import "time"

type TempCard struct {
	Id      int
	UserId  int
	CardTag string
	Issued  time.Time
}

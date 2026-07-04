package models

import "database/sql"

type Makerspace struct {
	Id          int
	Name        string
	Subtitle    string
	Description string
	DocsUrl     string
	ImageId     sql.NullInt32
	Hidden      bool
}

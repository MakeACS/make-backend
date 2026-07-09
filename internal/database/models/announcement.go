package models

type Announcement struct {
	Id           int
	Title        string
	Body         string
	MakerspaceId *int
}

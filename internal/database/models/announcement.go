package models

type Announcements struct {
	Id           int
	Title        string
	Body         string
	MakerspaceId *int
}

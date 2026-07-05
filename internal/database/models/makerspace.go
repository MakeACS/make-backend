package models

type Makerspace struct {
	Id          int
	Name        string
	Subtitle    string
	Description string
	DocsUrl     string
	ImageId     *int32
	Hidden      bool
	Timezone    string
}

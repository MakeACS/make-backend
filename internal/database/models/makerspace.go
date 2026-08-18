package models

type Makerspace struct {
	Id                 int
	Name               string
	Subtitle           string
	Description        string
	DocsUrl            string
	ImageId            *int32
	Hidden             bool
	Timezone           string
	ManagementAgroupId int
	StaffAgroupId      int // TODO: finer grained permissions
}

func (m Makerspace) LogEntity() LogEntity {
	return LogEntity{
		Id:    m.Id,
		Label: m.Name,
	}
}

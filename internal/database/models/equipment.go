package models

type Equipment struct {
	Id                      int
	Name                    string
	SubName                 string
	ZoneId                  int
	Hidden                  bool
	ImageId                 *int
	SopUrl                  string
	SignOffUrl              string
	Description             string
	Reservable              bool
	ReservationOnly         bool
	ReservationInstructions string
	NeedsWelcome            bool
	RequiresInPerson        bool
	RequiresTrainer         bool
}

func (e Equipment) LogEntity() LogEntity {
	return LogEntity{
		Id:    e.Id,
		Label: e.Name,
	}
}

type EquipmentInstance struct {
	Id              int
	EquipmentId     int
	Name            string
	AccessChannelId *int
}

func (e EquipmentInstance) LogEntity() LogEntity {
	return LogEntity{
		Id:    e.Id,
		Label: e.Name,
	}
}

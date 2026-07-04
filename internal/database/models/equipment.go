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

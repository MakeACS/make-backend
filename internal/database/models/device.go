package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Device struct {
	Id             int
	Name           string
	SN             string
	Paired         time.Time
	Hardware       *string
	Firmware       *string
	TargetFirmware *string
	KeyCycle       int
	MakerspaceId   int
}

type AccessDeviceFlags struct {
	LockWhenIdle      bool `json:"lock_when_idle"`
	RestartWhenUnused bool `json:"restart_when_unused"`
	Welcoming         bool `json:"welcoming"`
}

func (f *AccessDeviceFlags) Scan(value any) error {
	if value == nil {
		*f = AccessDeviceFlags{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for AccessDeviceFlags")
	}

	return json.Unmarshal(bytes, f)
}

func (f AccessDeviceFlags) Value() (driver.Value, error) {
	return json.Marshal(f)
}

type AccessComponentType int

const (
	Legacy         AccessComponentType = 0b000
	Switch1Channel AccessComponentType = 0b001
	Switch2Channel AccessComponentType = 0b010
	Switch3Channel AccessComponentType = 0b011
	Switch4Channel AccessComponentType = 0b100
	NonSwitching   AccessComponentType = 0b101
	InternalHMI    AccessComponentType = 0b110
	Communicative  AccessComponentType = 0b111
)

type AccessComponent struct {
	SN       string              `json:"SN"`
	Type     AccessComponentType `json:"type"`
	Children []AccessComponent   `json:"children"`
}

type AccessDeployment struct {
	SN         string            `json:"SN"`
	Components []AccessComponent `json:"components"`
}

func (d *AccessDeployment) Scan(value any) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for AccessDeployment")
	}

	return json.Unmarshal(bytes, d)
}

func (d AccessDeployment) Value() (driver.Value, error) {
	return json.Marshal(d)
}

type AccessDevice struct {
	DeviceId           int
	Channels           int
	TempDuration       int
	CurrentCardTag     string
	LastStatus         *time.Time
	SessionStart       *time.Time
	Flags              AccessDeviceFlags
	SealedDeployment   *AccessDeployment
	ReportedDeployment *AccessDeployment
}

type DispenserError string

const (
	CardStuck  DispenserError = "CARD_STUCK"
	OutOfCards DispenserError = "OUT_OF_CARDS"
)

type Dispenser struct {
	DeviceId  int
	CardsLeft int
	Error     *DispenserError
}

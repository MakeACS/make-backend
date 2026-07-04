package models

type AccessChannelState string

const (
	Idle      AccessChannelState = "IDLE"
	Unlocked  AccessChannelState = "UNLOCKED"
	AlwaysOn  AccessChannelState = "ALWAYS_ON"
	LockedOut AccessChannelState = "LOCKED_OUT"
	Fault     AccessChannelState = "FAULT"
)

type AccessChannel struct {
	Id           int
	DeviceId     int
	ChannelId    int
	State        *AccessChannelState
	TempDuration int
}

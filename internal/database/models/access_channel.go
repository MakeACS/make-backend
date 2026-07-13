package models

type AccessChannelState string

const (
	StateIdle      AccessChannelState = "IDLE"
	StateUnlocked  AccessChannelState = "UNLOCKED"
	StateAlwaysOn  AccessChannelState = "ALWAYS_ON"
	StateLockedOut AccessChannelState = "LOCKED_OUT"
	StateFault     AccessChannelState = "FAULT"
)

type AccessChannel struct {
	Id           int
	DeviceId     int
	ChannelId    int
	State        *AccessChannelState
	TempDuration int
}

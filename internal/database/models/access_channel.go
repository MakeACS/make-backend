package models

type AccessChannelState string

const (
	AccessChannelStateIdle      AccessChannelState = "IDLE"
	AccessChannelStateUnlocked  AccessChannelState = "UNLOCKED"
	AccessChannelStateAlwaysOn  AccessChannelState = "ALWAYS_ON"
	AccessChannelStateLockedOut AccessChannelState = "LOCKED_OUT"
	AccessChannelStateFault     AccessChannelState = "FAULT"
)

type AccessChannel struct {
	Id           int
	DeviceId     int
	ChannelId    int
	State        *AccessChannelState
	TempDuration int
}

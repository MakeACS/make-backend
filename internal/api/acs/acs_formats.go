package acs

import "make-backend/internal/database/models"

type AccessControllerState string

var (
	AccessControllerState_Idle      = "IDLE"
	AccessControllerState_Unlocked  = "UNLOCKED"
	AccessControllerState_AlwaysOn  = "ALWAYS_ON"
	AccessControllerState_LockedOut = "LOCKED_OUT"
	AccessControllerState_Fault     = "FAULT"
	AccessControllerState_Welcoming = "WELCOMING"
)

type ChannelState struct {
	ChannelID int64                 `json:"channelID"`
	State     AccessControllerState `json:"state"`
}
type AccessDeviceStatusReport struct {
	Channels       []ChannelState `json:"channels"`
	CurrentCardTag string         `json:"currentCardTag"`
}

type AccessDeviceStateChangeReason string

var (
	AccessDeviceStateChangeReason_Authed        = "AUTHED"
	AccessDeviceStateChangeReason_OverTemp      = "OVER_TEMP"
	AccessDeviceStateChangeReason_CardRemoved   = "CARD_REMOVED"
	AccessDeviceStateChangeReason_Commanded     = "COMMANDED"
	AccessDeviceStateChangeReason_Local         = "LOCAL"
	AccessDeviceStateChangeReason_IntegrityFail = "INTEGRITY_FAIL"
	AccessDeviceStateChangeReason_Fault         = "FAULT"
)

type AccessDeviceStateChangeReportChannel struct {
	ChannelID int64                         `json:"channelID"`
	FromState AccessControllerState         `json:"fromState"`
	ToState   AccessControllerState         `json:"toState"`
	Reason    AccessDeviceStateChangeReason `json:"reason"`
}

type AccessDeviceStateChangeReport struct {
	Channels       []AccessDeviceStateChangeReportChannel
	CurrentCardTag string
}

type AccessDeviceLogRequest struct {
	AuditLog bool   `json:"auditLog"`
	Message  string `json:"message"`
	Category string `json:"category,omitempty"`
}

type AccessDeviceAuthToRequest struct {
	State     AccessControllerState `json:"state"`
	CardTagId string                `json:"cardTagID"`
}
type AccessDeviceFlags struct {
	LockWhenIdle      bool `json:"lockWhenIdle"`
	RestartWhenUnused bool `json:"restartWhenUnused"`
	Welcoming         bool `json:"welcoming"`
}

type AccessDeviceConfigReportChannel struct {
	ChannelID    int64 `json:"channelID"`
	TempDuration int64 `json:"tempDurations"`
}

type AccessDeviceConfigReport struct {
	Channels   []AccessDeviceConfigReport   `json:"channels,omitempty"`
	InputMode  models.AccessDeviceInputMode `json:"inputMode,omitempty"`
	Deployment *ACSDeployment               `json:"deployment"`
	Flags      AccessDeviceFlags            `json:"flags"`
	Firmware   string                       `json:"firmware"`
}

type AccessDeviceInfoOptions string

var (
	AccessDeviceInfoOptions_Time  = "TIME"  // Current time
	AccessDeviceInfoOptions_State = "STATE" // State the channels should be in
	AccessDeviceInfoOptions_Hmi   = "HMI"   // Information intended for human consumption
	AccessDeviceInfoOptions_Flags = "FLAGS"
)

type AccessDeviceInfoRequest struct {
	Fields []AccessDeviceInfoOptions `json:"fields"`
}

type ServerAuthToChannel struct {
	ChannelID int64                 `json:"channelID"`
	State     AccessControllerState `json:"state"`
	Approved  bool                  `json:"approved"`
	Reason    string                `json:"reason"`
}

// Shape of what the server will send to the access device in response to an authTo request
type ServerAuthToResponse struct {
	Channels  []ServerAuthToChannel `json:"channels"`
	CardTagID string                `json:"cardTagID"`
}

type AccessDeviceFile string

var (
	AccessDeviceFile_Cert        = "CERT"
	AccessDeviceFile_OfflineList = "OFFLINE_LIST"
	AccessDeviceFile_Ota         = "OTA"
)

type ServerConfigUpdateRequestChannel struct {
	ID           int64              `json:"id"`
	TempDuration int64              `json:"tempDuration"`
	GetFiles     []AccessDeviceFile `json:"getFiles"`
}

/**
 * Shape of what the server sends to the access device when the server
 * wants the device to update its configuration
 */
type ServerConfigUpdateRequest struct {
	InputMode *models.AccessDeviceInputMode      `json:"inputMode"`
	Channels  []ServerConfigUpdateRequestChannel `json:"channels"`
}

type AccessDeviceActions string

var (
	AccessDeviceActions_Restart          = "RESTART"
	AccessDeviceActions_Seal             = "SEAL"
	AccessDeviceActions_Identify         = "IDENTIFY"
	AccessDeviceActions_ScheduledRestart = "SCHEDULED_RESTART"
)

// Shape of what the server sends to the access device when the server wants to command the device to take some action

type ServerCommand struct {
	ToState []struct {
		id    int64
		state AccessControllerState
	}
	action          *AccessDeviceActions
	identifyChannel int64
	flags           *AccessDeviceFlags
}

type AccessDeviceRole string

var (
	AccessDeviceRole_Welcome   = "WELCOME"
	AccessDeviceRole_Equipment = "EQUIPMENT"
)

type ServerStateResponse struct {
	ID    int64                 `json:"id"`
	State AccessControllerState `json:"state"`
}
type ServerHMIResponse struct {
	Role       AccessDeviceRole `json:"role"`
	Makerspace string           `json:"makerspace"`
	Channels   []struct {
		channelID    int64
		pairedEntity string
	} `json:"channels"`
}

// Shape of what the server sends the access device in response to an info request
type ServerInfoResponse struct {
	Time  int64                  `json:"time"` // milliseconds
	State *[]ServerStateResponse `json:"state"`
	Hmi   *[]ServerHMIResponse   `json:"hmi"`
	Flags *AccessDeviceFlags     `json:"flags"`
}

type WelcomeRequest struct {
	CardTagID string `json:"cardTagID"`
}

type WelcomeResponse struct {
	Welcomed  bool   `json:"welcomed"`
	CardTagID string `json:"cardTagID"`
}

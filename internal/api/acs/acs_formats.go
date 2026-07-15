package acs

import "make-backend/internal/database/models"

type ChannelState struct {
	ChannelID int                       `json:"channelID"`
	State     models.AccessChannelState `json:"state"`
}
type AccessDeviceStatusReport struct {
	Channels       []ChannelState `json:"channels"`
	CurrentCardTag string         `json:"currentCardTag"`
}

type AccessDeviceStateChangeReason string

var (
	AccessDeviceStateChangeReason_Authed        AccessDeviceStateChangeReason = "AUTHED"
	AccessDeviceStateChangeReason_OverTemp      AccessDeviceStateChangeReason = "OVER_TEMP"
	AccessDeviceStateChangeReason_CardRemoved   AccessDeviceStateChangeReason = "CARD_REMOVED"
	AccessDeviceStateChangeReason_Commanded     AccessDeviceStateChangeReason = "COMMANDED"
	AccessDeviceStateChangeReason_Local         AccessDeviceStateChangeReason = "LOCAL"
	AccessDeviceStateChangeReason_IntegrityFail AccessDeviceStateChangeReason = "INTEGRITY_FAIL"
	AccessDeviceStateChangeReason_Fault         AccessDeviceStateChangeReason = "FAULT"
)

type AccessChannelStateChangeReport struct {
	ChannelID int                           `json:"channelID"`
	FromState models.AccessChannelState     `json:"fromState"`
	ToState   models.AccessChannelState     `json:"toState"`
	Reason    AccessDeviceStateChangeReason `json:"reason"`
}

type AccessDeviceStateChangeReport struct {
	Channels       []AccessChannelStateChangeReport `json:"channels"`
	CurrentCardTag string                           `json:"currentCardTag"`
}

type AccessDeviceLogRequest struct {
	AuditLog bool   `json:"auditLog"`
	Message  string `json:"message"`
	Category string `json:"category,omitempty"`
}

type AccessDeviceAuthToRequest struct {
	State     models.AccessChannelState `json:"state"`
	CardTagId string                    `json:"cardTagID"`
}
type AccessDeviceFlags struct {
	LockWhenIdle      bool `json:"lockWhenIdle"`
	RestartWhenUnused bool `json:"restartWhenUnused"`
	Welcoming         bool `json:"welcoming"`
}

type AccessDeviceConfigReportChannel struct {
	ChannelID    int   `json:"channelID"`
	TempDuration int64 `json:"tempDuration"`
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
	AccessDeviceInfoOptions_Time         AccessDeviceInfoOptions = "TIME"  // Current time
	AccessDeviceInfoOptions_State        AccessDeviceInfoOptions = "STATE" // State the channels should be in
	AccessDeviceInfoOptions_Hmi          AccessDeviceInfoOptions = "HMI"   // Information intended for human consumption
	AccessDeviceInfoOptions_Flags        AccessDeviceInfoOptions = "FLAGS"
	AccessDeviceInfoOptions_HobbsTime    AccessDeviceInfoOptions = "HOBBS_TIME"
	AccessDeviceInfoOptions_CustomConfig AccessDeviceInfoOptions = "CUSTOM_CONFIG"
)

type AccessDeviceInfoRequest struct {
	Fields []AccessDeviceInfoOptions `json:"fields"`
}

type ServerAuthToChannel struct {
	ChannelID int                       `json:"channelID"`
	State     models.AccessChannelState `json:"state"`
	Approved  bool                      `json:"approved"`
	Reason    string                    `json:"reason"`
}

// Shape of what the server will send to the access device in response to an authTo request
type ServerAuthToResponse struct {
	Channels  []ServerAuthToChannel `json:"channels"`
	CardTagID string                `json:"cardTagID"`
}

type AccessDeviceFile string

var (
	AccessDeviceFile_Cert        AccessDeviceFile = "CERT"
	AccessDeviceFile_OfflineList AccessDeviceFile = "OFFLINE_LIST"
	AccessDeviceFile_Ota         AccessDeviceFile = "OTA"
)

type ServerConfigUpdateRequestChannel struct {
	ID           int   `json:"id"`
	TempDuration int64 `json:"tempDuration"`
}

/**
 * Shape of what the server sends to the access device when the server
 * wants the device to update its configuration
 */
type ServerConfigUpdateRequest struct {
	InputMode *models.AccessDeviceInputMode      `json:"inputMode"`
	Channels  []ServerConfigUpdateRequestChannel `json:"channels"`
	GetFiles  []AccessDeviceFile                 `json:"getFiles"`
}

type AccessDeviceActions string

var (
	AccessDeviceActions_Restart          AccessDeviceActions = "RESTART"
	AccessDeviceActions_Seal             AccessDeviceActions = "SEAL"
	AccessDeviceActions_Identify         AccessDeviceActions = "IDENTIFY"
	AccessDeviceActions_ScheduledRestart AccessDeviceActions = "SCHEDULED_RESTART"
)

// Shape of what the server sends to the access device when the server wants to command the device to take some action

type ServerCommand struct {
	ToState []struct {
		Id    int                       `json:"id"`
		State models.AccessChannelState `json:"state"`
	} `json:"toState"`
	Action          *AccessDeviceActions `json:"action"`
	IdentifyChannel int                  `json:"identifyChannel"`
	Flags           *AccessDeviceFlags   `json:"flags"`
}

type AccessDeviceRole string

var (
	AccessDeviceRole_Welcome   AccessDeviceRole = "WELCOME"
	AccessDeviceRole_Equipment AccessDeviceRole = "EQUIPMENT"
)

type ServerStateResponse struct {
	ID    int                       `json:"id"`
	State models.AccessChannelState `json:"state"`
}
type ServerHMIResponse struct {
	Role       AccessDeviceRole `json:"role"`
	Makerspace string           `json:"makerspace"`
	Channels   []struct {
		channelID    int
		pairedEntity string
	} `json:"channels"`
}

// Shape of what the server sends the access device in response to an info request
type ServerInfoResponse struct {
	Time      *int64                 `json:"time,omitempty"` // milliseconds
	State     *[]ServerStateResponse `json:"state,omitempty"`
	Hmi       *[]ServerHMIResponse   `json:"hmi,omitempty"`
	Flags     *AccessDeviceFlags     `json:"flags,omitempty"`
	HobbsTime *int64                 `json:"hobbs_time,omitempty"`
}

type WelcomeRequest struct {
	CardTagID string `json:"cardTagID"`
}

type WelcomeResponse struct {
	Welcomed  bool   `json:"welcomed"`
	CardTagID string `json:"cardTagID"`
}

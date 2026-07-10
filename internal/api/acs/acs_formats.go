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
type CoreStatusReport struct {
	Channels       []ChannelState `json:"channels"`
	CurrentCardTag string         `json:"currentCardTag"`
}

type CoreStateChangeReason string

var (
	CoreStateChangeReason_Authed        = "AUTHED"
	CoreStateChangeReason_OverTemp      = "OVER_TEMP"
	CoreStateChangeReason_CardRemoved   = "CARD_REMOVED"
	CoreStateChangeReason_Commanded     = "COMMANDED"
	CoreStateChangeReason_Local         = "LOCAL"
	CoreStateChangeReason_IntegrityFail = "INTEGRITY_FAIL"
	CoreStateChangeReason_Fault         = "FAULT"
)

type CoreStateChangeReportChannel struct {
	ChannelID int64                 `json:"channelID"`
	FromState AccessControllerState `json:"fromState"`
	ToState   AccessControllerState `json:"toState"`
	Reason    CoreStateChangeReason `json:"reason"`
}

type CoreStateChangeReport struct {
	Channels       []CoreStateChangeReportChannel
	CurrentCardTag string
}

type CoreLogRequest struct {
	AuditLog bool   `json:"auditLog"`
	Message  string `json:"message"`
	Category string `json:"category,omitempty"`
}

type CoreAuthToRequest struct {
	State     AccessControllerState `json:"state"`
	CardTagId string                `json:"cardTagID"`
}
type CoreFlags struct {
	LockWhenIdle      bool `json:"lockWhenIdle"`
	RestartWhenUnused bool `json:"restartWhenUnused"`
	Welcoming         bool `json:"welcoming"`
}

type CoreConfigReportChannel struct {
	ChannelID    int64 `json:"channelID"`
	TempDuration int64 `json:"tempDurations"`
}

type CoreConfigReport struct {
	Channels   []CoreConfigReport   `json:"channels,omitempty"`
	InputMode  models.CoreInputMode `json:"inputMode,omitempty"`
	Deployment *ACSDeployment       `json:"deployment"`
	Flags      CoreFlags            `json:"flags"`
	Firmware   string               `json:"firmware"`
}

type CoreInfoOptions string

var (
	CoreInfoOptions_Time  = "TIME"  // Current time
	CoreInfoOptions_State = "STATE" // State the channels should be in
	CoreInfoOptions_Hmi   = "HMI"   // Information intended for human consumption
	CoreInfoOptions_Flags = "FLAGS"
)

type CoreInfoRequest struct {
	Fields []CoreInfoOptions `json:"fields"`
}

type ServerAuthToChannel struct {
	ChannelID int64                 `json:"channelID"`
	State     AccessControllerState `json:"state"`
	Approved  bool                  `json:"approved"`
	Reason    string                `json:"reason"`
}

/**
 * Shape of what the server will send to the core in response to
 * an authTo request
 */
type ServerAuthToResponse struct {
	Channels  []ServerAuthToChannel `json:"channels"`
	CardTagID string                `json:"cardTagID"`
}

type CoreFile string

var (
	CoreFile_Cert        = "CERT"
	CoreFile_OfflineList = "OFFLINE_LIST"
	CoreFile_Ota         = "OTA"
)

type ServerConfigUpdateRequestChannel struct {
	ID           int64      `json:"id"`
	TempDuration int64      `json:"tempDuration"`
	GetFiles     []CoreFile `json:"getFiles"`
}

/**
 * Shape of what the server sends to the core when the server
 * wants the core to update its configuration
 */
type ServerConfigUpdateRequest struct {
	InputMode *models.CoreInputMode              `json:"inputMode"`
	Channels  []ServerConfigUpdateRequestChannel `json:"channels"`
}

type CoreActions string

var (
	CoreActions_Restart          = "RESTART"
	CoreActions_Seal             = "SEAL"
	CoreActions_Identify         = "IDENTIFY"
	CoreActions_ScheduledRestart = "SCHEDULED_RESTART"
)

// Shape of what the server sends to the core when the server
// wants to command the core to take some action

type ServerCommand struct {
	ToState []struct {
		id    int64
		state AccessControllerState
	}
	action          *CoreActions
	identifyChannel int64
	flags           *CoreFlags
}

type CoreRole string

var (
	CoreRole_Welcome   = "WELCOME"
	CoreRole_Equipment = "EQUIPMENT"
)

type ServerStateResponse struct {
	ID    int64                 `json:"id"`
	State AccessControllerState `json:"state"`
}
type ServerHMIResponse struct {
	Role       CoreRole `json:"role"`
	Makerspace string   `json:"makerspace"`
	Channels   []struct {
		channelID    int64
		pairedEntity string
	} `json:"channels"`
}

// Shape of what the server sends the core in response to
// an info request
type ServerInfoResponse struct {
	Time  int64                  `json:"time"` // milliseconds
	State *[]ServerStateResponse `json:"state"`
	Hmi   *[]ServerHMIResponse   `json:"hmi"`
	Flags *CoreFlags             `json:"flags"`
}

type WelcomeRequest struct {
	CardTagID string `json:"cardTagID"`
}

type WelcomeResponse struct {
	Welcomed  bool   `json:"welcomed"`
	CardTagID string `json:"cardTagID"`
}

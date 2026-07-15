package acs

type ACSDeployment struct {
	SN         string // Access Device Serial Number
	Components []ACSComponent
}

type ACSComponentType uint

var (
	LEGACY           ACSComponentType = 0b000
	SWITCH_1_CHANNEL                  = 0b001
	SWITCH_2_CHANNEL                  = 0b010
	SWITCH_3_CHANNEL                  = 0b011
	SWITCH_4_CHANNEL                  = 0b100
	NON_SWITCHING                     = 0b101
	INTERNAL_HMI                      = 0b110
	COMMUNICATIVE                     = 0b111
)

type ACSComponent struct {
	SN            string // Component Serial Number
	ComponentType ACSComponentType
	Children      []ACSComponent
}

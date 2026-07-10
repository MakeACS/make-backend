package acs

import "make-backend/internal/database"

type ACSOrchestrator struct {
	store *database.Store
}

func (orc *ACSOrchestrator) HandleSendAccessDeviceCommand() {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) HandleWelcomeRequest() {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) CommandAllAccessDevices() {
	panic("unimplemented")
}

func (orc *ACSOrchestrator) CommandMakerspaceAccessDevices() {
	panic("unimplemented")

}

package models

type RunStatus string

const (
	Completed RunStatus = "COMPLETED"
	Running   RunStatus = "IN_PROGRESS"
	Failed    RunStatus = "FAILED"
	Submitted RunStatus = "SUBMITTED"
)

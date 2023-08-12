package models

import "time"

type StartExport struct {
	LastRunTime time.Time `json:"lastRunTime"`
	RunStatus   RunStatus `json:"runStatus"`
	TaskToken   string    `json:"taskToken"`
}

type ExportStatus struct {
	DatastoreId string    `json:"datastoreId"`
	JobStatus   RunStatus `json:"jobStatus"`
	JobId       string    `json:"jobId"`
}

type ExportDescription struct {
	JobProperties ExportStatus   `json:"exportJobProperties"`
	Output        []ExportOutput `json:"output"`
}

type ExportStatusWithTaskToken struct {
	ExportStatus
	TaskToken string `json:"taskToken"`
}

type ExportOutput struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

package models

type StartExportRequest struct {
	JobName           string           `json:"JobName"`
	OutputDataConfig  OutputDataConfig `json:"OutputDataConfig"`
	DataAccessRoleArn string           `json:"DataAccessRoleArn"`
}

type S3Configuration struct {
	S3Uri    string `json:"S3Uri"`
	KmsKeyId string `json:"KmsKeyId"`
}

type OutputDataConfig struct {
	S3Configuration S3Configuration `json:"S3Configuration"`
}

type StartExportResponse struct {
	DatastoreId string    `json:"datastoreId"`
	JobStatus   RunStatus `json:"jobStatus"`
	JobId       string    `json:"jobId"`
}

func NewStartExportRequest(jobName string, s3 string, keyId string, roleArn string) *StartExportRequest {
	return &StartExportRequest{
		JobName: jobName,
		OutputDataConfig: OutputDataConfig{
			S3Configuration: S3Configuration{
				S3Uri:    s3,
				KmsKeyId: keyId,
			},
		},
		DataAccessRoleArn: roleArn,
	}
}

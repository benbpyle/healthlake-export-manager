package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStartExportRequest(t *testing.T) {
	result := NewStartExportRequest("jobName", "s3", "keyId", "roleARN")
	assert.Equal(t, "jobName", result.JobName)
	assert.Equal(t, "s3", result.OutputDataConfig.S3Configuration.S3Uri)
	assert.Equal(t, "keyId", result.OutputDataConfig.S3Configuration.KmsKeyId)
	assert.Equal(t, "roleARN", result.DataAccessRoleArn)
}

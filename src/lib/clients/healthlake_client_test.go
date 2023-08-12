package clients

import (
	"bytes"
	"cdc/lib/models"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/stretchr/testify/assert"
)

type MockHttpClient struct {
	IHealthLakeClient
	ResponseType string
	IsSuccess    bool
}

func (m MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if !m.IsSuccess {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"error": true}`)),
		}, nil
	}
	if m.ResponseType == "DescribeExport" {
		exportDescription := models.ExportDescription{
			JobProperties: models.ExportStatus{
				JobId: "1",
			},
		}
		body, _ := json.Marshal(exportDescription)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	}
	if m.ResponseType == "StartExport" {
		exportResponse := models.StartExportResponse{
			JobId: "1",
		}
		body, _ := json.Marshal(exportResponse)
		return &http.Response{
			StatusCode: http.StatusAccepted,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil

	}
	return nil, errors.New("bad response type")
}

func getClient(responseType string, isSuccess bool) *HealthLakeClient {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	signer := v4.NewSigner(sess.Config.Credentials)
	client := MockHttpClient{
		ResponseType: responseType,
		IsSuccess:    isSuccess,
	}
	return NewHealthLakeClient(client, signer, "endpoint", "abcd1234", "us-west-2")
}

func TestNewHttpClient(t *testing.T) {
	NewHttpClient()
}

func TestHealthLakeClientDescribeExport(t *testing.T) {
	client := getClient("DescribeExport", true)
	exportDescription, err := client.DescribeExport(context.TODO(), "abcd")
	assert.NoError(t, err)
	assert.Equal(t, "1", exportDescription.JobProperties.JobId)
}

func TestHealthLakeClientStartExport(t *testing.T) {
	client := getClient("StartExport", true)
	exportResponse, err := client.StartExport(context.TODO(), time.Now(), "s3://test", "kmsKeyId", "role")
	assert.NoError(t, err)
	assert.Equal(t, "1", exportResponse.JobId)
}

func TestHealthLakeClientStartExportError(t *testing.T) {
	client := getClient("StartExport", false)
	_, err := client.StartExport(context.TODO(), time.Now(), "s3://test", "kmsKeyId", "role")
	assert.Error(t, err)
}

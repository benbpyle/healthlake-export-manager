package main

import (
	"cdc/lib/clients"
	"cdc/lib/models"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

type MockS3Client struct {
	clients.IS3Client
}

func (MockS3Client) DownloadFile(ctx context.Context, bucket string, event *models.ExportOutput) (*string, error) {
	return aws.String("file"), nil
}

func TestHandler(t *testing.T) {
	s3Client = MockS3Client{}
	parseFile = func(string, string) ([]models.CurantisPublishedEvent, error) {
		return []models.CurantisPublishedEvent{}, nil
	}
	deleteFile = func(string) error {
		return nil
	}

	event, err := handler(context.TODO(), &models.ExportOutput{})
	assert.NoError(t, err)
	assert.NotNil(t, event)
}

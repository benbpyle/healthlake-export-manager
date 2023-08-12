package main

import (
	"cdc/lib/clients"
	"cdc/lib/models"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

var (
	event = events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
				Body:      `{"lastRunTime":"2023-08-09T18:45:34.478Z","runStatus":"COMPLETED","taskToken":"abcd1234"}`,
			},
		},
	}
)

type MockSQSClient struct {
	clients.ISQSClient
	IsSuccess *bool
}

func (m MockSQSClient) SendReCheckMessage(context.Context, *models.ExportStatusWithTaskToken) (*string, error) {
	if m.IsSuccess != nil && !*m.IsSuccess {
		return nil, errors.New("SQS error")
	}
	return aws.String("abcd1234"), nil
}

type MockHealthlakeClient struct {
	clients.IHealthLakeClient
	IsSuccess *bool
}

func (m MockHealthlakeClient) StartExport(context.Context, time.Time, string, string, string) (*models.StartExportResponse, error) {
	if m.IsSuccess != nil && !*m.IsSuccess {
		return nil, errors.New("SQS error")
	}
	return &models.StartExportResponse{}, nil
}

func TestHandler(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{}
	sqsClient = MockSQSClient{}
	err := handler(context.TODO(), event)
	assert.NoError(t, err)
}

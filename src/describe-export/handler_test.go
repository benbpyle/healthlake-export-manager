package main

import (
	"cdc/lib/clients"
	"cdc/lib/models"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/stretchr/testify/assert"
)

var (
	event = events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
				Body:      "{ \"datastoreId\": \"465cbdb19feb7e747fefb74a2cd0c428\", \"jobId\": \"b05cd350a0230a2eabc0fada7b77ab18\", \"jobStatus\": \"SUBMITTED\", \"taskToken\": \"AQB8AAAAKgAAAAMAAAAAAAAAAeVGELEdP/DkYwcw2GToscGZ+xwia6YwbliYRVJnPGaMSKfaaQ58fSJwyCx/9P1TfUT04eon2mLsZJfqY0Vut2jLtaSTSH+jn2VzKw==k8luxbuIH6Oj5OGmmmI0yAdoo1TYPH9soDSQPm2ZmgZAd50OCfOsXAPJDzw88ZpTP2W8sz3i9ulWaARgh17AH763hw760Xfrx3MH4Px+NRpw0IWWxYS4tEVrxbamCbkiV1PPnc1xhmq/3FjCgOe+zo6fytPOSsSew7aQGD/0wjIXXmlnQTtMFnYLoiTZvcXSQTVzVr/0sdt/jKY3OcEi3WGIv0nLRZQ94bT+UFsb0w4dqYkCOjFasmS0+azkdlLlOohznCD511sD5+LTG7sziJLLp0Tuvsifn/XF5jv4diEGuGNT9SG9x5PSCdTyHyuziLtUay0F3LnTBurMxYUBjg1OMxVUpiTfXViJE1OeAQd0+IS6FnJc0QbHHTF3Ib03481m6OCNw+xMKwt/R0QNO8Lck3lfk0DFzF04kA7VNq4F674okFdru4yzeJaIBJnpeMNb3UYk/sHl67TuxiPupS+LzDdwi1C6JCRoIo6ap26ZJklz/OJgxq5ZZW/a9+DeXO+Lmq7jqiJbb+Q3hkGm\" }",
			},
		},
	}
)

type MockHealthlakeClient struct {
	clients.IHealthLakeClient
	Result models.RunStatus
}

func (m MockHealthlakeClient) DescribeExport(ctx context.Context, jobId string) (*models.ExportDescription, error) {
	switch m.Result {
	case models.Completed:
		return &models.ExportDescription{
			JobProperties: models.ExportStatus{
				JobStatus: models.Completed,
			},
		}, nil
	case models.Submitted:
		return &models.ExportDescription{
			JobProperties: models.ExportStatus{
				JobStatus: models.Submitted,
			},
		}, nil
	case models.Failed:
		return &models.ExportDescription{
			JobProperties: models.ExportStatus{
				JobStatus: models.Failed,
			},
		}, nil
	}
	return nil, errors.New("export error")
}

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

type MockSFNClient struct {
}

func (MockSFNClient) SendTaskFailure(context.Context, *sfn.SendTaskFailureInput, ...func(*sfn.Options)) (*sfn.SendTaskFailureOutput, error) {
	return &sfn.SendTaskFailureOutput{}, nil
}
func (MockSFNClient) SendTaskSuccess(context.Context, *sfn.SendTaskSuccessInput, ...func(*sfn.Options)) (*sfn.SendTaskSuccessOutput, error) {
	return &sfn.SendTaskSuccessOutput{}, nil
}
func (MockSFNClient) SendTaskHeartbeat(context.Context, *sfn.SendTaskHeartbeatInput, ...func(*sfn.Options)) (*sfn.SendTaskHeartbeatOutput, error) {
	return &sfn.SendTaskHeartbeatOutput{}, nil
}

type MockS3Client struct {
	clients.IS3Client
	IsSuccess *bool
}

func (m MockS3Client) UploadManifest(ctx context.Context, bucket string, export *models.ExportDescription) (*string, error) {
	if m.IsSuccess != nil && !*m.IsSuccess {
		return nil, errors.New("S3 error")
	}
	return aws.String("file"), nil
}

func TestHandlerCompleted(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{
		Result: models.Completed,
	}
	sqsClient = MockSQSClient{}
	sfnClient = MockSFNClient{}
	s3Client = MockS3Client{}

	err := handler(context.TODO(), event)
	assert.NoError(t, err)
}

func TestHandleSubmitted(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{
		Result: models.Submitted,
	}
	sqsClient = MockSQSClient{}
	sfnClient = MockSFNClient{}
	s3Client = MockS3Client{}

	err := handler(context.TODO(), event)
	assert.NoError(t, err)
}

func TestHandleFailed(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{
		Result: models.Failed,
	}
	sqsClient = MockSQSClient{}
	sfnClient = MockSFNClient{}
	s3Client = MockS3Client{}

	err := handler(context.TODO(), event)
	assert.NoError(t, err)
}

func TestHandleHealthlakeError(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{}
	sqsClient = MockSQSClient{}
	sfnClient = MockSFNClient{}
	s3Client = MockS3Client{}

	err := handler(context.TODO(), event)
	assert.Error(t, err)
}

func TestHandlerSQSError(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{
		Result: models.Submitted,
	}
	sqsClient = MockSQSClient{
		IsSuccess: aws.Bool(false),
	}
	sfnClient = MockSFNClient{}
	s3Client = MockS3Client{}

	err := handler(context.TODO(), event)
	assert.Error(t, err)
}

func TestHandlerS3Error(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{
		Result: models.Completed,
	}
	sqsClient = MockSQSClient{}
	sfnClient = MockSFNClient{}
	s3Client = MockS3Client{
		IsSuccess: aws.Bool(false),
	}

	err := handler(context.TODO(), event)
	assert.Error(t, err)
}

func TestHandlerBadMessage(t *testing.T) {
	healthLakeClient = MockHealthlakeClient{
		Result: models.Submitted,
	}
	sqsClient = MockSQSClient{}
	sfnClient = MockSFNClient{}
	s3Client = MockS3Client{}

	err := handler(context.TODO(),
		events.SQSEvent{
			Records: []events.SQSMessage{
				{
					MessageId: "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
					Body:      "bad",
				},
			},
		})
	assert.Error(t, err)
}

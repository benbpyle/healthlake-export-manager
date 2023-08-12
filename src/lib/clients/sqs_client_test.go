package clients

import (
	"cdc/lib/models"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
)

type MockAWSSqsClient struct {
	IAWSSqsClient
}

func (MockAWSSqsClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return &sqs.SendMessageOutput{
		MessageId: aws.String("abcd1234"),
	}, nil
}

func TestNewSQSClient(t *testing.T) {
	NewSQSClient(aws.NewConfig(), "recheckUrl")
}

func TestSqsClientSendReCheckMessage(t *testing.T) {
	sqsClient := SqsClient{
		Client: MockAWSSqsClient{},
	}
	result, err := sqsClient.SendReCheckMessage(context.TODO(), &models.ExportStatusWithTaskToken{})
	assert.NoError(t, err)
	assert.Equal(t, "abcd1234", *result)
}

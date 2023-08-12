package clients

import (
	"cdc/lib/models"
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type ISQSClient interface {
	SendReCheckMessage(ctx context.Context, m *models.ExportStatusWithTaskToken) (*string, error)
}

type IAWSSqsClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type SqsClient struct {
	ISQSClient
	Client          IAWSSqsClient
	reCheckQueueUrl string
}

// NewStepFunctionsClient inits a StepFunctionsClient session to be used throughout the services
func NewSQSClient(cfg *aws.Config, reCheckQueueUrl string) *SqsClient {
	client := sqs.NewFromConfig(*cfg)

	return &SqsClient{
		Client:          client,
		reCheckQueueUrl: reCheckQueueUrl,
	}
}

func (s *SqsClient) SendReCheckMessage(ctx context.Context, m *models.ExportStatusWithTaskToken) (*string, error) {
	b, err := json.Marshal(m)
	sb := string(b)
	if err != nil {
		return nil, err
	}

	message := sqs.SendMessageInput{
		QueueUrl:     &s.reCheckQueueUrl,
		DelaySeconds: 30,
		MessageBody:  &sb,
	}

	output, err := s.Client.SendMessage(ctx, &message)

	if err != nil {
		return nil, err
	}

	return output.MessageId, nil
}

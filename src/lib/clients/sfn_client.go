package clients

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

type ISFNClient interface {
	SendTaskFailure(ctx context.Context, params *sfn.SendTaskFailureInput, optFns ...func(*sfn.Options)) (*sfn.SendTaskFailureOutput, error)
	SendTaskSuccess(ctx context.Context, params *sfn.SendTaskSuccessInput, optFns ...func(*sfn.Options)) (*sfn.SendTaskSuccessOutput, error)
	SendTaskHeartbeat(ctx context.Context, params *sfn.SendTaskHeartbeatInput, optFns ...func(*sfn.Options)) (*sfn.SendTaskHeartbeatOutput, error)
}

type SFNClient struct {
	*sfn.Client
}

// NewStepFunctionsClient inits a StepFunctionsClient session to be used throughout the services
func NewStepFunctionsClient(cfg *aws.Config) SFNClient {
	client := SFNClient{
		Client: sfn.NewFromConfig(*cfg),
	}

	return client
}

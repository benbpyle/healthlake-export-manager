package clients

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestNewStepFunctionsClient(t *testing.T) {
	assert.NotNil(t, NewStepFunctionsClient(&aws.Config{}))
}

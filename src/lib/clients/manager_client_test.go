package clients

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var (
	client = NewS3Client(aws.NewConfig()).client
)

func TestManagerNewUploader(t *testing.T) {
	m := Manager{}
	m.NewUploader(client)
}

func TestManagerNewDownloader(t *testing.T) {
	m := Manager{}
	m.NewDownloader(client)
}

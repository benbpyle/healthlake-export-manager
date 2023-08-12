package clients

import (
	"cdc/lib/models"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
)

type MockIO struct {
}

func (MockIO) Open(name string) (IFile, error)   { return &MockFile{}, nil }
func (MockIO) Create(name string) (IFile, error) { return &MockFile{}, nil }
func (MockIO) Remove(name string) error          { return nil }

type MockFile struct {
	index int
}

func (*MockFile) Write(p []byte) (n int, err error)              { return 0, nil }
func (*MockFile) WriteAt(p []byte, off int64) (n int, err error) { return 0, nil }
func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.index == 2 {
		err = io.EOF
	}
	test := "{ \"resourceType\": \"Patient\" }\n"
	m.index = m.index + 1
	n = copy(p, []byte(test))
	return
}
func (*MockFile) Close() error { return nil }

type MockManager struct {
}

func (MockManager) NewUploader(client manager.UploadAPIClient) IUploader  { return MockUploader{} }
func (MockManager) NewDownloader(c manager.DownloadAPIClient) IDownloader { return MockDownloader{} }

type MockUploader struct {
}

func (MockUploader) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	return &manager.UploadOutput{}, nil
}

type MockDownloader struct {
}

func (MockDownloader) Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(*manager.Downloader)) (n int64, err error) {
	return 0, nil
}

func TestNewS3Client(t *testing.T) {
	client := NewS3Client(&aws.Config{})
	assert.NotNil(t, client)
}

func TestCreateFile(t *testing.T) {
	filesystem = MockIO{}
	path, err := createFile(&models.ExportDescription{})
	assert.NoError(t, err)
	assert.NotNil(t, path)
}

func TestS3ClientUploadManifest(t *testing.T) {
	filesystem = MockIO{}
	localManager = MockManager{}
	client := NewS3Client(&aws.Config{})

	result, err := client.UploadManifest(context.TODO(), "testBucket", &models.ExportDescription{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestS3ClientDownloadFile(t *testing.T) {
	t.Skip() // TODO
	filesystem = MockIO{}
	localManager = MockManager{}
	client := NewS3Client(&aws.Config{})

	result, err := client.DownloadFile(context.TODO(), "testBucket", &models.ExportOutput{
		Url: "https://healthlake.us-west-2.amazonaws.com/datastore/1997a81c16b83d7f77dfdd555e322f86/r4/$export?_since=2023-07-18T00%3A00%3A00.461Z",
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDeleteFile(t *testing.T) {
	filesystem = MockIO{}
	err := DeleteFile("patient.json")
	assert.NoError(t, err)
}

func TestParseFileSuccess(t *testing.T) {
	filesystem = MockIO{}
	localManager = MockManager{}
	newCurantisPublishedEvent = func(interface{}, string) (*models.CurantisPublishedEvent, error) {
		return &models.CurantisPublishedEvent{}, nil
	}
	events, err := ParseFile("patient.json", "Patient")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(events))
}

func TestParseFileError(t *testing.T) {
	filesystem = MockIO{}
	localManager = MockManager{}
	newCurantisPublishedEvent = func(interface{}, string) (*models.CurantisPublishedEvent, error) {
		return &models.CurantisPublishedEvent{}, errors.New("error")
	}
	_, err := ParseFile("patient.json", "Patient")
	assert.NoError(t, err) // this is a warning
}

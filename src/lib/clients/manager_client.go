package clients

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type IManager interface {
	NewUploader(client manager.UploadAPIClient) IUploader
	NewDownloader(c manager.DownloadAPIClient) IDownloader
}

type IUploader interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

type IDownloader interface {
	Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(*manager.Downloader)) (n int64, err error)
}

type Manager struct {
}

func (Manager) NewUploader(client manager.UploadAPIClient) IUploader {
	return manager.NewUploader(client)
}

func (Manager) NewDownloader(client manager.DownloadAPIClient) IDownloader {
	return manager.NewDownloader(client)
}

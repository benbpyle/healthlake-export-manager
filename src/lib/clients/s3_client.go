package clients

import (
	"bufio"
	"cdc/lib/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/segmentio/ksuid"
	"github.com/sirupsen/logrus"
)

var (
	filesystem                IFilesystem = OsFS{}
	localManager              IManager    = Manager{}
	newCurantisPublishedEvent             = models.NewCurantisPublishedEvent
)

type IS3Client interface {
	UploadManifest(ctx context.Context, bucket string, export *models.ExportDescription) (*string, error)
	DownloadFile(ctx context.Context, bucket string, event *models.ExportOutput) (*string, error)
}

type S3Client struct {
	IS3Client
	client *s3.Client
}

func NewS3Client(cfg *aws.Config) *S3Client {
	client := s3.NewFromConfig(*cfg)

	return &S3Client{
		client: client,
	}
}

func createFile(export *models.ExportDescription) (*string, error) {
	filePath := fmt.Sprintf("/tmp/%s.json", ksuid.New().String())

	bytes, err := json.Marshal(export.Output)
	if err != nil {
		return nil, err
	}

	f, err := filesystem.Create(filePath)

	if err != nil {
		return nil, err
	}

	_, err = f.Write(bytes)

	if err != nil {
		return nil, err
	}

	defer f.Close()
	return &filePath, nil
}

func (s *S3Client) UploadManifest(ctx context.Context, bucket string, export *models.ExportDescription) (*string, error) {
	filePath, err := createFile(export)
	if err != nil {
		return nil, err
	}
	o, _ := filesystem.Open(*filePath)
	defer o.Close()

	key := fmt.Sprintf("%s/%s-FHIR_EXPORT-%s/manifest.json", "exports", export.JobProperties.DatastoreId, export.JobProperties.JobId)
	input := &s3.PutObjectInput{
		Bucket:      &bucket,
		Body:        o,
		Key:         &key,
		ContentType: aws.String("application/json"),
	}

	uploader := localManager.NewUploader(s.client)

	uploadResp, err := uploader.Upload(ctx, input)

	logrus.WithFields(logrus.Fields{
		"a":   uploadResp,
		"err": err,
	}).Debug("Uploaded results")

	return &key, err
}

func (s *S3Client) DownloadFile(ctx context.Context, bucket string, event *models.ExportOutput) (*string, error) {
	filePath := fmt.Sprintf("/tmp/%s.json", ksuid.New().String())
	file, err := filesystem.Create(filePath)

	if err != nil {
		return nil, err
	}

	downloader := localManager.NewDownloader(s.client)
	urlParts := strings.Split(event.Url, "exports")

	if len(urlParts) != 2 {
		return nil, fmt.Errorf("incoming uri is in the incorrect format")
	}

	key := fmt.Sprintf("exports%s", urlParts[1])
	input := &s3.GetObjectInput{
		Key:    &key,
		Bucket: &bucket,
	}

	_, err = downloader.Download(ctx, file, input)

	if err != nil {
		return nil, err
	}

	return &filePath, nil
}

func DeleteFile(fileName string) error {
	err := filesystem.Remove(fileName)

	if err != nil {
		return err
	}

	return nil
}

func ParseFile(fileName string, t string) ([]models.CurantisPublishedEvent, error) {
	var publishers []models.CurantisPublishedEvent

	jsonFile, err := filesystem.Open(fileName)

	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	fileScanner := bufio.NewScanner(jsonFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		// fileLines = append(fileLines, fileScanner.Text())
		line := fileScanner.Text()

		var i interface{}
		err := json.Unmarshal([]byte(line), &i)
		if err != nil {
			return nil, err
		}
		var event *models.CurantisPublishedEvent
		event, err = newCurantisPublishedEvent(i, t)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Warn("error extracting metadata")
		}
		publishers = append(publishers, *event)
	}

	return publishers, nil
}

package main

import (
	"cdc/lib/clients"
	"cdc/lib/models"
	"cdc/lib/util"
	"context"
	"strconv"

	"os"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/aws/aws-lambda-go/lambda"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"

	log "github.com/sirupsen/logrus"
)

var (
	isLocal    bool
	s3Client   clients.IS3Client
	bucketName string
	parseFile  = clients.ParseFile
	deleteFile = clients.DeleteFile
)

func handler(ctx context.Context, event *models.ExportOutput) ([]models.CurantisPublishedEvent, error) {
	log.WithFields(log.Fields{
		"body": event,
	}).Debug("Printing out the body")

	f, err := s3Client.DownloadFile(ctx, bucketName, event)

	if err != nil {
		return nil, err
	}

	cpe, err := parseFile(*f, event.Type)
	deleteFile(*f)
	return cpe, err
}

func main() {
	lambda.Start(ddlambda.WrapFunction(handler, util.DataDogConfig()))
}

func init() {
	isLocal, _ = strconv.ParseBool(os.Getenv("IS_LOCAL"))

	log.SetFormatter(&log.JSONFormatter{
		PrettyPrint: isLocal,
	})

	bucketName = os.Getenv("BUCKET")
	awsCfg, _ := awscfg.LoadDefaultConfig(context.Background())
	awstrace.AppendMiddleware(&awsCfg)
	s3Client = clients.NewS3Client(&awsCfg)
	util.SetLevel(os.Getenv("LOG_LEVEL"))
}

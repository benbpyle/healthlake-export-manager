package main

import (
	"cdc/lib/clients"
	"cdc/lib/models"
	"cdc/lib/util"
	"context"
	"encoding/json"
	"errors"
	"strconv"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"

	"os"

	log "github.com/sirupsen/logrus"
)

var isLocal bool
var healthLakeClient clients.IHealthLakeClient
var sqsClient clients.ISQSClient

func handler(ctx context.Context, event events.SQSEvent) error {
	if len(event.Records) != 1 {
		return errors.New("too many in the batch")
	}

	body := &models.StartExport{}
	err := json.Unmarshal([]byte(event.Records[0].Body), body)

	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"body": body,
	}).Debug("Printing out the body")

	keyArn := os.Getenv("KEY_ARN")
	roleArn := os.Getenv("ROLE_ARN")
	s3Uri := os.Getenv("S3_URI")

	exportOutput, err := healthLakeClient.StartExport(ctx, body.LastRunTime, s3Uri, keyArn, roleArn)

	if err != nil {
		return err
	}

	messageOutput, err := sqsClient.SendReCheckMessage(ctx, &models.ExportStatusWithTaskToken{
		ExportStatus: models.ExportStatus{
			DatastoreId: exportOutput.DatastoreId,
			JobStatus:   exportOutput.JobStatus,
			JobId:       exportOutput.JobId,
		}, TaskToken: body.TaskToken,
	})

	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"exportOutput":  exportOutput,
		"messageOutput": messageOutput,
	}).Debug("Printing out the send success output")

	return nil
}

func main() {
	lambda.Start(ddlambda.WrapFunction(handler, util.DataDogConfig()))
}

func init() {
	isLocal, _ = strconv.ParseBool(os.Getenv("IS_LOCAL"))

	log.SetFormatter(&log.JSONFormatter{
		PrettyPrint: isLocal,
	})

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})

	if err != nil {
		log.Fatalf("failed creating session: %s", err)
	}

	signer := v4.NewSigner(sess.Config.Credentials)
	awsCfg, _ := awscfg.LoadDefaultConfig(context.Background())
	awstrace.AppendMiddleware(&awsCfg)

	httpClient := clients.NewHttpClient()
	sqsClient = clients.NewSQSClient(&awsCfg, os.Getenv("RECHECK_QUEUE_URL"))
	healthLakeClient = clients.NewHealthLakeClient(httpClient, signer, os.Getenv("HEALTHLAKE_ENDPOINT"), os.Getenv("HEALTHLAKE_DATASTOREID"), os.Getenv("HEALTHLAKE_REGION"))
	util.SetLevel(os.Getenv("LOG_LEVEL"))
}

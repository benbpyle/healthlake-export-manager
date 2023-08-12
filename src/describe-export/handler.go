package main

import (
	"cdc/lib/clients"
	"cdc/lib/models"
	"cdc/lib/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	ddlambda "github.com/DataDog/datadog-lambda-go"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
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
var sfnClient clients.ISFNClient
var s3Client clients.IS3Client
var bucketName string

func handler(ctx context.Context, event events.SQSEvent) error {
	if len(event.Records) != 1 {
		return errors.New("too many in the batch")
	}

	exd := &models.ExportStatusWithTaskToken{}
	err := json.Unmarshal([]byte(event.Records[0].Body), exd)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("error unmarshalling SQS record.  Moving along")
		return err
	}

	log.WithFields(log.Fields{
		"body": exd,
	}).Debug("Printing out the body")

	exportOutput, err := healthLakeClient.DescribeExport(ctx, exd.JobId)

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("error describing job")

		input := &sfn.SendTaskFailureInput{
			TaskToken: &exd.TaskToken,
		}

		_, _ = sfnClient.SendTaskFailure(ctx, input)

		return err
	}

	log.WithFields(log.Fields{
		"healthLakeStatus": exportOutput,
	}).Debug("results from the describe")
	if exportOutput.JobProperties.JobStatus == models.Completed {
		file, err := s3Client.UploadManifest(ctx, bucketName, exportOutput)

		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("error uploading manifest")
			return err
		}

		strOutput := fmt.Sprintf("{\"bucket\": \"%s\", \"manifest\": \"%s\"}", bucketName, *file)
		input := &sfn.SendTaskSuccessInput{
			TaskToken: &exd.TaskToken,
			Output:    &strOutput,
		}

		_, err = sfnClient.SendTaskSuccess(ctx, input)

		// need to fail the run and put the message back on the queue
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("error sending task success")
			return err
		}

		log.WithFields(log.Fields{
			"jobId": exportOutput.JobProperties.JobId,
		}).Debug("Wrapping up with with a completed status")
	} else if exportOutput.JobProperties.JobStatus == models.Submitted ||
		exportOutput.JobProperties.JobStatus == models.Running {
		input := &sfn.SendTaskHeartbeatInput{
			TaskToken: &exd.TaskToken,
		}

		_, err := sfnClient.SendTaskHeartbeat(ctx, input)

		// need to fail the run and put the message back on the queue
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("error sending task heartbeat")
			return err
		}

		messageOutput, err := sqsClient.SendReCheckMessage(ctx, &models.ExportStatusWithTaskToken{
			ExportStatus: models.ExportStatus{
				DatastoreId: exd.DatastoreId,
				JobStatus:   exportOutput.JobProperties.JobStatus,
				JobId:       exd.JobId,
			}, TaskToken: exd.TaskToken,
		})

		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("error putting message back on queue after heartbeat")
			return err
		}

		log.WithFields(log.Fields{
			"exportOutput":  exportOutput,
			"messageOutput": messageOutput,
		}).Debug("Printing out the putting message back on queue")

	} else if exportOutput.JobProperties.JobStatus == models.Failed {
		input := &sfn.SendTaskFailureInput{
			TaskToken: &exd.TaskToken,
		}

		_, err := sfnClient.SendTaskFailure(ctx, input)

		// need to fail the run and put the message back on the queue
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("error sending task failure")
			return err
		}

		log.WithFields(log.Fields{
			"jobId": exportOutput.JobProperties.JobId,
		}).Debug("Wrapping up with with a failed status")
	}

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

	bucketName = os.Getenv("BUCKET")
	awsCfg, _ := awscfg.LoadDefaultConfig(context.Background())
	awstrace.AppendMiddleware(&awsCfg)
	signer := v4.NewSigner(sess.Config.Credentials)
	httpClient := clients.NewHttpClient()
	sqsClient = clients.NewSQSClient(&awsCfg, os.Getenv("RECHECK_QUEUE_URL"))
	sfnClient = clients.NewStepFunctionsClient(&awsCfg)
	s3Client = clients.NewS3Client(&awsCfg)
	healthLakeClient = clients.NewHealthLakeClient(httpClient, signer, os.Getenv("HEALTHLAKE_ENDPOINT"), os.Getenv("HEALTHLAKE_DATASTOREID"), os.Getenv("HEALTHLAKE_REGION"))
	util.SetLevel(os.Getenv("LOG_LEVEL"))
}

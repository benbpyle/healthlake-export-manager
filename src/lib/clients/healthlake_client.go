package clients

import (
	"cdc/lib/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/sirupsen/logrus"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
)

type IHealthLakeClient interface {
	DescribeExport(ctx context.Context, jobId string) (*models.ExportDescription, error)
	StartExport(ctx context.Context, t time.Time, s3Uri string, kmsKeyId string, role string) (*models.StartExportResponse, error)
}

type IHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HealthLakeClient struct {
	IHealthLakeClient
	httpClient            IHttpClient
	signer                *v4.Signer
	healthLakeEndpoint    string
	healthLakeDatastoreId string
	healthLakeRegion      string
}

func NewHttpClient() *http.Client {
	client := httptrace.WrapClient(&http.Client{}, httptrace.RTWithResourceNamer(func(h *http.Request) string {
		return fmt.Sprintf("%s %s://%s%s", h.Method, h.URL.Scheme, h.URL.Host, h.URL.Path)
	}))

	return client
}

func NewHealthLakeClient(
	httpClient IHttpClient,
	signer *v4.Signer,
	healthLakeEndpoint string,
	healthLakeDatastoreId string,
	healthLakeRegion string) *HealthLakeClient {
	return &HealthLakeClient{
		httpClient:            httpClient,
		healthLakeEndpoint:    healthLakeEndpoint,
		healthLakeDatastoreId: healthLakeDatastoreId,
		healthLakeRegion:      healthLakeRegion,
		signer:                signer,
	}
}

func (hc *HealthLakeClient) DescribeExport(ctx context.Context, jobId string) (*models.ExportDescription, error) {
	url := fmt.Sprintf("https://%s/%s/r4/export/%s", hc.healthLakeEndpoint, hc.healthLakeDatastoreId, jobId)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	logrus.Debugf("(URL)=%s", url)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	_, err = hc.signer.Sign(req, nil, "healthlake", hc.healthLakeRegion, time.Now())

	if err != nil {
		return nil, err
	}

	resp, err := hc.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("error describing export job (%s) in healthlake with code (%d)", jobId, resp.StatusCode)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var js models.ExportDescription
	err = decoder.Decode(&js)

	if err != nil {
		return nil, err
	}

	return &js, nil
}

func (hc *HealthLakeClient) StartExport(ctx context.Context, t time.Time, s3Uri string, kmsKeyId string, role string) (*models.StartExportResponse, error) {
	formatted := t.Format(time.RFC3339)
	url := fmt.Sprintf("https://%s/%s/r4/$export?_since=%s", hc.healthLakeEndpoint, hc.healthLakeDatastoreId, formatted)
	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Debug("What's going to run")
	b, _ := json.Marshal(models.NewStartExportRequest(
		formatted,
		s3Uri,
		kmsKeyId,
		role,
	))
	body := strings.NewReader(string(b))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/fhir+json")
	req.Header.Set("Prefer", "respond-async")
	_, err = hc.signer.Sign(req, body, "healthlake", hc.healthLakeRegion, time.Now())

	if err != nil {
		return nil, err
	}

	resp, err := hc.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 202 {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		var js interface{}
		_ = decoder.Decode(&js)
		logrus.WithFields(logrus.Fields{
			"body": js,
		}).Error("error starting job")
		return nil, fmt.Errorf("error starting export job in healthlake with code (%d)", resp.StatusCode)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var js models.StartExportResponse
	err = decoder.Decode(&js)

	if err != nil {
		return nil, err
	}

	return &js, nil
}

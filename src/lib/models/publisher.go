package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/bix-digital/golang-fhir-models/fhir-models/fhir"
	"github.com/segmentio/ksuid"
)

var (
	baseIdentifierUrl = "https://fhir.curantissolutions.com/identifier"
)

type CurantisPublishedEvent struct {
	MetaDetails MetaDetails `json:"meta"`
	Details     interface{} `json:"data"`
}

type MetaDetails struct {
	CorrelationId              string    `json:"correlationId"`
	SourceTime                 time.Time `json:"sourceTime"`
	Source                     string    `json:"source"`
	Version                    string    `json:"version"`
	TenantId                   *int      `json:"tenantId,omitempty"`
	IdentifierQueryStringValue *string   `json:"identifierQueryStringValue,omitempty"`
}

type CustomResource struct {
	fhir.Resource
	ResourceType string            `json:"resourceType"`
	Identifier   []fhir.Identifier `json:"identifier,omitempty"`
	Extension    []fhir.Extension  `json:"extension,omitempty"`
}

func NewCurantisPublishedEvent(contents interface{}, t string) (*CurantisPublishedEvent, error) {
	m := MetaDetails{
		CorrelationId: ksuid.New().String(),
		SourceTime:    time.Now(),
		Source:        fmt.Sprintf("com.curantis.fhir-%s", strings.ToLower(t)),
		Version:       "2.0",
	}

	err := extractMetadata(contents, &m)

	return &CurantisPublishedEvent{
		MetaDetails: m,
		Details:     contents,
	}, err
}

func extractMetadata(contents interface{}, metaDetails *MetaDetails) error {
	contentsBytes, err := json.Marshal(contents)
	parseError := false
	if err != nil {
		parseError = true
	}
	resource := &CustomResource{}
	err = json.Unmarshal(contentsBytes, resource)
	if err != nil {
		parseError = true
	}

	metaDetails.TenantId = getTenantId(resource.Extension)
	metaDetails.IdentifierQueryStringValue = getIdentifierQueryStringValue(resource)
	if parseError {
		return errors.New("error parsing fhir resource")
	}
	if metaDetails.TenantId == nil || metaDetails.IdentifierQueryStringValue == nil {
		return errors.New("either tenantId or identifierQueryStringValue could not be resolved from the resource")
	}
	return nil
}

func getTenantId(extensions []fhir.Extension) *int {
	for _, extension := range extensions {
		if extension.Url == "https://fhir.curantissolutions.com/ext/StructureDefinition/AssociatedOrganization" {
			tenantId, err := strconv.Atoi(*extension.ValueString)
			if err == nil {
				return &tenantId
			}
		}
	}
	return nil
}

func getIdentifierQueryStringValue(resource *CustomResource) *string {
	for _, identifier := range resource.Identifier {
		if (identifier.System != nil && identifier.Value != nil) &&
			*identifier.System == fmt.Sprintf("%s/%s", baseIdentifierUrl, resource.ResourceType) {
			return aws.String(fmt.Sprintf("%s/%s|%s", baseIdentifierUrl, resource.ResourceType, *identifier.Value))
		}
	}
	return nil
}

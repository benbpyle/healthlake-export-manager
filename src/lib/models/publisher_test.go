package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCurantisPublishedEvent(t *testing.T) {
	var i interface{}
	content := `{"authoredOn":"2023-08-09T19:52:35.064Z","contained":[{"code":{"coding":[{"display":"Abacavir Sulfate Oral Solution 20 MG/ML","system":"https://www.wolterskluwer.com"}]},"id":"med","ingredient":[{"itemCodeableConcept":{"text":"ingredient0"},"strength":{"denominator":{"code":"ML","system":"http://unitsofmeasure.org","unit":"ML","value":1},"numerator":{"code":"MG","system":"http://unitsofmeasure.org","unit":"MG","value":20}}}],"resourceType":"Medication"}],"dosageInstruction":[{"doseAndRate":[{"doseQuantity":{"unit":"Solution"},"type":{"coding":[{"code":"ordered"}]}}],"route":{"text":"Oral"},"text":"2 times per day","timing":{"event":["2023-08-09T04:00:00Z"]}}],"extension":[{"id":"associated-organization","url":"https://fhir.curantissolutions.com/ext/StructureDefinition/AssociatedOrganization","valueString":"697190"}],"id":"697190-10336","identifier":[{"id":"697190-10336","system":"https://fhir.curantissolutions.com/identifier/MedicationRequest","use":"official","value":"697190-10336"}],"intent":"order","medicationReference":{"reference":"#med"},"requester":{"identifier":{"system":"https://fhir.curantissolutions.com/identifier/MedicationRequest","value":"726569"}},"resourceType":"MedicationRequest","status":"active","subject":{"identifier":{"system":"https://fhir.curantissolutions.com/identifier/Patient","value":"697190-000000000203277"}},"meta":{"lastUpdated":"2023-08-09T19:52:43.705Z"}}`
	err := json.Unmarshal([]byte(content), &i)
	assert.NoError(t, err)
	event, err := NewCurantisPublishedEvent(i, "Patient")
	assert.NoError(t, err)
	assert.Equal(t, 697190, *event.MetaDetails.TenantId)
}

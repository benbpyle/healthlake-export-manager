package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataDogConfig(t *testing.T) {
	assert.Equal(t, false, DataDogConfig().DebugLogging)
}

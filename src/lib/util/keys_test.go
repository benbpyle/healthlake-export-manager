package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrelationIdKeyString(t *testing.T) {
	var key CorrelationIdKey = "abc"
	assert.Equal(t, "abc", key.String())
}

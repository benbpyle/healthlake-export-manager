package util

import (
	"testing"
)

func TestSetLevel(t *testing.T) {
	SetLevel("error")
	SetLevel("info")
	SetLevel("debug")
	SetLevel("other")
	SetLevel("trace")
}

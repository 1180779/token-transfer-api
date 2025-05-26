package tests

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// This test fails to check the workflow.
func TestFail(t *testing.T) {
	assert.Equal(t, true, false, "")
}

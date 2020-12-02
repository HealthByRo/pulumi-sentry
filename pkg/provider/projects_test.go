package provider

import (
	"testing"

	"github.com/stvp/assert"
)

func TestOk(t *testing.T) {
	assert.True(t, true)
}

func TestFail(t *testing.T) {
	assert.True(t, false)
}

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRuntimeMetrics(t *testing.T) {
	metrics := getRuntimeMetrics()
	assert.NotEmpty(t, metrics)
}

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRuntimeMetrics(t *testing.T) {
	metrics, err := getRuntimeMetrics()
	assert.NotEmpty(t, metrics)
	assert.NoError(t, err)
}

func TestGetAdditionalRuntimeMetrics(t *testing.T) {
	metrics, err := getAdditionalMetrics()
	assert.NotEmpty(t, metrics)
	assert.NoError(t, err)
}

package service

import (
	"math/rand"
	"testing"

	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetRuntimeMetrics(t *testing.T) {
	metrics := getRuntimeMetrics()
	assert.NotEmpty(t, metrics)
}

func TestUpdateRuntimeMetrics(t *testing.T) {

	testData := []struct {
		name                  string
		inputMap              []map[string]float64
		expectedUpdateCounter int64
	}{
		{
			name: "Одно значение без изменения",
			inputMap: []map[string]float64{
				{"test1": 1.0, "test2": rand.Float64()},
				{"test1": 1.0, "test3": rand.Float64()}},
			expectedUpdateCounter: 3,
		},
		{
			name: "Все значения обновлены",
			inputMap: []map[string]float64{
				{"test1": rand.Float64(), "test2": rand.Float64()},
				{"test1": rand.Float64(), "test3": rand.Float64()}},
			expectedUpdateCounter: 4,
		},
	}

	agentConfig := config.AgentConfig{
		RemoteAddr: "localhost:8080",
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			collector := BuildCollector(agentConfig)
			var counter int64
			for _, testMap := range test.inputMap {
				counter += collector.UpdateMetrics(testMap)
			}
			assert.Equal(t, test.expectedUpdateCounter, counter)
		})
	}
}

package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateGauge(t *testing.T) {
	memStorage := BuildMemStorage()

	testData := []struct {
		name          string
		gauge         *float64
		expectedExist bool
	}{
		{
			name:          "test1",
			gauge:         nil,
			expectedExist: true,
		},
		{
			name:          "test2",
			gauge:         new(float64),
			expectedExist: true,
		},
		{
			name: "test3",
			gauge: func(f float64) *float64 {
				return &f
			}(3.14),
			expectedExist: true,
		},
		{
			name: "test4",
			gauge: func(f float64) *float64 {
				return &f
			}(-3.14),
			expectedExist: true,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			memStorage.UpdateGauge(data.name, data.gauge)
			val, isExist := memStorage.GetGauge(data.name)

			assert.Equal(t, data.expectedExist, isExist)
			assert.Equal(t, data.gauge, val)
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	memStorage := BuildMemStorage()

	testData := []struct {
		name          string
		counter       *int64
		expectedExist bool
	}{
		{
			name:          "test1",
			counter:       nil,
			expectedExist: true,
		},
		{
			name:          "test2",
			counter:       new(int64),
			expectedExist: true,
		},
		{
			name: "test3",
			counter: func(f int64) *int64 {
				return &f
			}(314),
			expectedExist: true,
		},
		{
			name: "test4",
			counter: func(f int64) *int64 {
				return &f
			}(-314),
			expectedExist: true,
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			memStorage.UpdateCounter(data.name, data.counter)
			val, isExist := memStorage.GetCounter(data.name)

			assert.Equal(t, data.expectedExist, isExist)
			assert.Equal(t, data.counter, val)
		})
	}
}

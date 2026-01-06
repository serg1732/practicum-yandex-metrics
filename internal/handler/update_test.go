package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	handlerBuilder := BuildUpdateHandler(repository.BuildMemStorage())
	testData := []struct {
		name           string
		req            *http.Request
		expectedStatus int
	}{
		{
			name:           "test 1",
			req:            httptest.NewRequest("POST", fmt.Sprintf("/update/%f/%f/%f", rand.Float64(), rand.Float64(), rand.Float64()), nil),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test 2",
			req:            httptest.NewRequest("POST", fmt.Sprintf("/update/%s/%s/%v", "gauge", "metric", rand.Float64()), nil),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test 2",
			req:            httptest.NewRequest("POST", fmt.Sprintf("/update/%s/%s/%v", "counter", "metric", rand.Int63()), nil),
			expectedStatus: http.StatusOK,
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", handlerBuilder.UpdateHandler)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, td.req)
			assert.Equal(t, td.expectedStatus, rr.Code)
		})
	}
}

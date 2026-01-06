package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestSuccessServerHandler(t *testing.T) {
	storage := repository.BuildMemStorage()
	srv := httptest.NewServer(buildHttpMux(handler.BuildUpdateHandler(storage)))
	defer srv.Close()

	testData := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "test 1",
			method:         "POST",
			url:            fmt.Sprintf("%s/update/%f/%f/%f", srv.URL, rand.Float64(), rand.Float64(), rand.Float64()),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test 2",
			method:         "POST",
			url:            fmt.Sprintf("%s/update/%f/%f", srv.URL, rand.Float64(), rand.Float64()),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "test 3",
			method:         "POST",
			url:            fmt.Sprintf("%s/update/%s/%s/%f", srv.URL, "gauge", "metricName", rand.Float64()),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test 4",
			method:         "POST",
			url:            fmt.Sprintf("%s/update/%s/%s/%d", srv.URL, "counter", "metricName", rand.Int64()),
			expectedStatus: http.StatusOK,
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			req, errBuildRequest := http.NewRequest(td.method, td.url, nil)
			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			assert.Nil(t, errBuildRequest)
			assert.Equal(t, td.expectedStatus, resp.StatusCode)
		})
	}
}

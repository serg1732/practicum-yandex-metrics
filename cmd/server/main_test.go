package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUpdateServerHandler(t *testing.T) {
	storage := repository.BuildMemStorage(context.Background(),
		&config.ServerConfig{
			RunAddr:         "127.0.0.1:8080",
			StoreInternal:   0,
			FileStoragePath: "storage.json",
			Restore:         false,
		})
	updateHandler := handler.BuildUpdateHandler(storage)
	readHandlers := handler.BuildReadHandler(storage)

	srv := httptest.NewServer(buildRouter(slog.Default(), updateHandler, readHandlers))
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

func TestAllReadServerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMemStorage(ctrl)

	updateHandler := handler.BuildUpdateHandler(mockRepo)
	readHandlers := handler.BuildReadHandler(mockRepo)

	srv := httptest.NewServer(buildRouter(slog.Default(), updateHandler, readHandlers))
	defer srv.Close()

	testData := []struct {
		name            string
		expectedStatus  int
		expectedCounter map[string]*models.Metrics
		expectedGauges  map[string]*models.Metrics
	}{
		{"test 1",
			http.StatusOK,
			map[string]*models.Metrics{
				"test-counter": {
					ID:    "test-counter",
					MType: models.Counter,
					Delta: getPtr(rand.Int64()),
				},
			},
			map[string]*models.Metrics{
				"test-gauge": {
					ID:    "test-gauge",
					MType: models.Gauge,
					Value: getPtr(rand.Float64()),
				},
			},
		},
		{"test 2",
			http.StatusOK,
			map[string]*models.Metrics{
				"test-counter": {
					ID:    "test-counter",
					MType: models.Counter,
					Delta: getPtr(rand.Int64()),
				},
			},
			map[string]*models.Metrics{},
		},
		{"test 3",
			http.StatusOK,
			map[string]*models.Metrics{},
			map[string]*models.Metrics{
				"test-gauge": {
					ID:    "test-gauge",
					MType: models.Gauge,
					Value: getPtr(rand.Float64()),
				}},
		},
		{"test 4",
			http.StatusOK,
			map[string]*models.Metrics{},
			map[string]*models.Metrics{},
		},
		{"test 5",
			http.StatusOK,
			map[string]*models.Metrics{},
			nil,
		},
		{"test 6",
			http.StatusOK,
			nil,
			map[string]*models.Metrics{},
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			mockRepo.
				EXPECT().
				GetAllCounters().
				Return(td.expectedCounter)
			mockRepo.
				EXPECT().
				GetAllGauges().
				Return(td.expectedGauges)
			resp, err := http.DefaultClient.Get(srv.URL)
			if err != nil {
				assert.Nil(t, err)
			}
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, td.expectedStatus, resp.StatusCode)
			expectedPage, err := getExpectedPage(td.expectedCounter, td.expectedGauges)
			assert.NoError(t, err)
			assert.Equal(t, string(expectedPage), string(respBody))
		})
	}
}

func TestSelectReadServerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMemStorage(ctrl)

	updateHandler := handler.BuildUpdateHandler(mockRepo)
	readHandlers := handler.BuildReadHandler(mockRepo)

	srv := httptest.NewServer(buildRouter(slog.Default(), updateHandler, readHandlers))
	defer srv.Close()

	testData := []struct {
		name             string
		url              string
		expectedStatus   int
		repoCounterKey   string
		repoCounterValue *models.Metrics
		repoCounterExist bool
		repoGaugesKey    string
		repoGaugesValue  *models.Metrics
		repoGaugesExist  bool
	}{
		{
			name:             "test 1 not found",
			url:              "/value/counter/test1",
			expectedStatus:   http.StatusNotFound,
			repoCounterKey:   "test1",
			repoCounterExist: false,
		},
		{
			name:            "test 2 not found",
			url:             "/value/gauge/test2",
			expectedStatus:  http.StatusNotFound,
			repoGaugesKey:   "test2",
			repoGaugesExist: false,
		},
		{
			name:           "test 3 bad type",
			url:            "/value/hunter/test",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "test 4",
			url:            "/value/counter/test4",
			expectedStatus: http.StatusOK,
			repoCounterKey: "test4",
			repoCounterValue: &models.Metrics{
				ID:    "test4",
				MType: models.Counter,
				Delta: getPtr(rand.Int64()),
			},
			repoCounterExist: true,
		},
		{
			name:           "test 5",
			url:            "/value/gauge/test5",
			expectedStatus: http.StatusOK,
			repoGaugesKey:  "test5",
			repoGaugesValue: &models.Metrics{
				ID:    "test5",
				MType: models.Gauge,
				Value: getPtr(rand.Float64()),
			},
			repoGaugesExist: true,
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			mockRepo.
				EXPECT().
				GetCounter(gomock.Eq(td.repoCounterKey)).
				Return(td.repoCounterValue, td.repoCounterExist).AnyTimes()
			mockRepo.
				EXPECT().
				GetGauge(gomock.Eq(td.repoGaugesKey)).
				Return(td.repoGaugesValue, td.repoGaugesExist).AnyTimes()

			resp, err := http.DefaultClient.Get(srv.URL + td.url)
			if err != nil {
				assert.Nil(t, err)
			}
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, td.expectedStatus, resp.StatusCode)
			if resp.StatusCode == http.StatusOK {
				if td.repoCounterExist {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoCounterValue.Delta), string(respBody))
				} else if td.repoGaugesExist {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoGaugesValue.Value), string(respBody))
				}
			}
		})
	}
}

func getPtr[T any](v T) *T {
	return &v
}

func getExpectedPage(counter map[string]*models.Metrics, gauge map[string]*models.Metrics) ([]byte, error) {
	templateHTML, errParseTemplate := template.New("All").Parse(getTemplate())
	if errParseTemplate != nil {
		return nil, errParseTemplate
	}

	data := struct {
		GaugeMap   map[string]*models.Metrics
		CounterMap map[string]*models.Metrics
	}{
		GaugeMap:   gauge,
		CounterMap: counter,
	}
	var buffer bytes.Buffer
	wr := bufio.NewWriter(&buffer)
	err := templateHTML.Execute(wr, data)
	if err != nil {
		return nil, err
	}
	wr.Flush()
	return buffer.Bytes(), nil
}

func getTemplate() string {
	return `
	<!doctype html>
	<html>
		<head>
    	<meta charset="utf-8">
    	<title>Metrics</title>
		</head>
		<body>
			<h3>gauge</h3>
				<pre>
					{{range $k, $v := .GaugeMap}}
					{{$k}}={{$v.Value}}
					{{end}}
				</pre>

			<h3>counter</h3>
		<pre>
			{{range $k, $v := .CounterMap}}
			{{$k}}={{$v.Delta}}
			{{end}}
		</pre>
		</body>
	</html>
`
}

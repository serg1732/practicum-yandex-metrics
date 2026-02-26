package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler/mocks"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestAllGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReadStorage(ctrl)

	handlerBuilder := BuildReadHandler(mockRepo)
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
					MType: "counter",
					Delta: getPtr(rand.Int64()),
				},
			},
			map[string]*models.Metrics{
				"test-gauge": {
					ID:    "test-gauge",
					MType: "gauge",
					Value: getPtr(rand.Float64()),
				},
			},
		},
		{"test 2",
			http.StatusOK,
			map[string]*models.Metrics{
				"test-counter": {
					ID:    "test-counter",
					MType: "counter",
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
					MType: "gauge",
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
				GetAllCounters(gomock.Any()).
				Return(td.expectedCounter, nil)
			mockRepo.
				EXPECT().
				GetAllGauges(gomock.Any()).
				Return(td.expectedGauges, nil)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /", handlerBuilder.AllMetricsHandler(slog.Default()))
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			mux.ServeHTTP(rr, req)

			assert.Equal(t, td.expectedStatus, rr.Code)
			expectedPage, err := getExpectedPage(td.expectedCounter, td.expectedGauges)
			assert.NoError(t, err)
			assert.Equal(t, string(expectedPage), rr.Body.String())
		})
	}
}

func TestSelectReadServerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockReadStorage(ctrl)
	handlerBuilder := BuildReadHandler(mockRepo)

	testData := []struct {
		name             string
		url              string
		err              error
		expectedStatus   int
		repoCounterKey   string
		repoCounterValue *models.Metrics
		repoGaugesKey    string
		repoGaugesValue  *models.Metrics
	}{
		{
			name:           "test 1 not found",
			url:            "/value/counter/test1",
			expectedStatus: http.StatusNotFound,
			repoCounterKey: "test1",
			err:            repository.ErrorMetricNotFound,
		},
		{
			name:           "test 2 not found",
			url:            "/value/gauge/test2",
			expectedStatus: http.StatusNotFound,
			repoGaugesKey:  "test2",
			err:            repository.ErrorMetricNotFound,
		},
		{
			name:           "test 3 bad type",
			url:            "/value/hunter/test",
			expectedStatus: http.StatusNotFound,
			err:            repository.ErrorMetricNotFound,
		},
		{
			name:           "test 4 success counter",
			url:            "/value/counter/test4",
			expectedStatus: http.StatusOK,
			repoCounterKey: "test4",
			repoCounterValue: &models.Metrics{
				ID:    "test4",
				MType: "counter",
				Delta: getPtr(rand.Int64()),
			},
		},
		{
			name:           "test 5 success gauges",
			url:            "/value/gauge/test5",
			expectedStatus: http.StatusOK,
			repoGaugesKey:  "test5",
			repoGaugesValue: &models.Metrics{
				ID:    "test5",
				MType: "gauge",
				Value: getPtr(rand.Float64()),
			},
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			mockRepo.
				EXPECT().
				GetCounter(gomock.Any(), gomock.Eq(td.repoCounterKey)).
				Return(td.repoCounterValue, td.err).AnyTimes()
			mockRepo.
				EXPECT().
				GetGauge(gomock.Any(), gomock.Eq(td.repoGaugesKey)).
				Return(td.repoGaugesValue, td.err).AnyTimes()

			r := chi.NewRouter()
			r.HandleFunc("GET /value/{metricType}/{metricName}", handlerBuilder.SelectMetricHandler(slog.Default()))
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", td.url, nil)
			r.ServeHTTP(rr, req)

			assert.Equal(t, td.expectedStatus, rr.Code)
			if rr.Code == http.StatusOK {
				if td.repoCounterValue != nil {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoCounterValue.Delta), rr.Body.String())
				} else if td.repoGaugesValue != nil {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoGaugesValue.Value), rr.Body.String())
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

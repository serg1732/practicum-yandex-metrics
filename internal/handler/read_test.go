package handler

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAllGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMemStorage(ctrl)

	handlerBuilder := BuildReadHandler(mockRepo)
	testData := []struct {
		name            string
		expectedStatus  int
		expectedCounter map[string]*int64
		expectedGauges  map[string]*float64
	}{
		{"test 1",
			http.StatusOK,
			map[string]*int64{
				"test-counter": getPtr(rand.Int64()),
			},
			map[string]*float64{
				"test-gauge": getPtr(rand.Float64()),
			},
		},
		{"test 2",
			http.StatusOK,
			map[string]*int64{
				"test-counter": getPtr(rand.Int64()),
			},
			map[string]*float64{},
		},
		{"test 3",
			http.StatusOK,
			map[string]*int64{},
			map[string]*float64{
				"test-gauge": getPtr(rand.Float64())},
		},
		{"test 4",
			http.StatusOK,
			map[string]*int64{},
			map[string]*float64{},
		},
		{"test 5",
			http.StatusOK,
			map[string]*int64{},
			nil,
		},
		{"test 6",
			http.StatusOK,
			nil,
			map[string]*float64{},
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

			mux := http.NewServeMux()
			mux.HandleFunc("GET /", handlerBuilder.AllMetricsHandler)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			mux.ServeHTTP(rr, req)

			assert.Equal(t, td.expectedStatus, rr.Code)
			assert.Equal(t, string(getExpectedPage(td.expectedCounter, td.expectedGauges)), rr.Body.String())

		})
	}
}

func TestSelectReadServerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMemStorage(ctrl)
	handlerBuilder := BuildReadHandler(mockRepo)

	testData := []struct {
		name             string
		url              string
		expectedStatus   int
		repoCounterKey   string
		repoCounterValue *int64
		repoCounterExist bool
		repoGaugesKey    string
		repoGaugesValue  *float64
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
			name:             "test 4 success counter",
			url:              "/value/counter/test4",
			expectedStatus:   http.StatusOK,
			repoCounterKey:   "test4",
			repoCounterValue: getPtr(rand.Int64()),
			repoCounterExist: true,
		},
		{
			name:            "test 5 success gauges",
			url:             "/value/gauge/test5",
			expectedStatus:  http.StatusOK,
			repoGaugesKey:   "test5",
			repoGaugesValue: getPtr(rand.Float64()),
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

			r := chi.NewRouter()
			r.HandleFunc("GET /value/{metricType}/{metricName}", handlerBuilder.SelectMetricHandler)
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", td.url, nil)
			r.ServeHTTP(rr, req)

			assert.Equal(t, td.expectedStatus, rr.Code)
			if rr.Code == http.StatusOK {
				if td.repoCounterExist {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoCounterValue), rr.Body.String())
				} else if td.repoGaugesExist {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoGaugesValue), rr.Body.String())
				}
			}
		})
	}
}

func getPtr[T any](v T) *T {
	return &v
}

func getExpectedPage(counter map[string]*int64, gauge map[string]*float64) []byte {
	w := new(bytes.Buffer)
	fmt.Fprintln(w, "<!doctype html><html><body>")
	fmt.Fprintln(w, "<h3>gauge</h3><pre>")
	for k, v := range gauge {
		fmt.Fprintf(w, "%s=%v\n", k, *v)
	}
	fmt.Fprintln(w, "</pre>")

	fmt.Fprintln(w, "<h3>counter</h3><pre>")
	for k, v := range counter {
		fmt.Fprintf(w, "%s=%v\n", k, *v)
	}
	fmt.Fprintln(w, "</pre></body></html>")
	return w.Bytes()
}

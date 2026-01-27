package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUpdateServerHandler(t *testing.T) {
	storage := repository.BuildMemStorage()
	updateHandler := handler.BuildUpdateHandler(storage)
	readHandlers := handler.BuildReadHandler(storage)

	srv := httptest.NewServer(buildRouter(updateHandler, readHandlers))
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

	srv := httptest.NewServer(buildRouter(updateHandler, readHandlers))
	defer srv.Close()

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
			resp, err := http.DefaultClient.Get(srv.URL)
			if err != nil {
				assert.Nil(t, err)
			}
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, td.expectedStatus, resp.StatusCode)
			assert.Equal(t, string(getExpectedPage(td.expectedCounter, td.expectedGauges)), string(respBody))
		})
	}
}

func TestSelectReadServerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockMemStorage(ctrl)

	updateHandler := handler.BuildUpdateHandler(mockRepo)
	readHandlers := handler.BuildReadHandler(mockRepo)

	srv := httptest.NewServer(buildRouter(updateHandler, readHandlers))
	defer srv.Close()

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
			name:             "test 4",
			url:              "/value/counter/test4",
			expectedStatus:   http.StatusOK,
			repoCounterKey:   "test4",
			repoCounterValue: getPtr(rand.Int64()),
			repoCounterExist: true,
		},
		{
			name:            "test 5",
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

			resp, err := http.DefaultClient.Get(srv.URL + td.url)
			if err != nil {
				assert.Nil(t, err)
			}
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, td.expectedStatus, resp.StatusCode)
			if resp.StatusCode == http.StatusOK {
				if td.repoCounterExist {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoCounterValue), string(respBody))
				} else if td.repoGaugesExist {
					assert.Equal(t, fmt.Sprintf("%v\n", *td.repoGaugesValue), string(respBody))
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

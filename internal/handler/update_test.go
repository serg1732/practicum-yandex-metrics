package handler

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/serg1732/practicum-yandex-metrics/internal/config"
	"github.com/serg1732/practicum-yandex-metrics/internal/handler/mocks"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuditor := mocks.NewMockAuditor(ctrl)

	handlerBuilder := BuildUpdateHandler(
		repository.BuildMemStorage(
			context.Background(),
			slog.Default(),
			&config.ServerConfig{},
		),
		mockAuditor,
	)

	testData := []struct {
		req            *http.Request
		name           string
		expectedStatus int
	}{
		{
			req: httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				"/update/unknown/metric/123",
				nil,
			),
			name:           "bad metric type",
			expectedStatus: http.StatusBadRequest,
		},
		{
			req: httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				"/update/gauge/metric/123.45",
				nil,
			),
			name:           "gauge ok",
			expectedStatus: http.StatusOK,
		},
		{
			req: httptest.NewRequestWithContext(
				t.Context(),
				http.MethodPost,
				"/update/counter/metric/10",
				nil,
			),
			name:           "counter ok",
			expectedStatus: http.StatusOK,
		},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			if td.expectedStatus == http.StatusOK {
				mockAuditor.EXPECT().
					BroadCast(gomock.Any(), gomock.Any()).
					Times(1)
			}

			mux := http.NewServeMux()

			mux.HandleFunc(
				"POST /update/{metricType}/{metricName}/{metricValue}",
				handlerBuilder.UpdatePathValuesHandler(slog.Default()),
			)

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, td.req)

			assert.Equal(t, td.expectedStatus, rr.Code)
		})
	}
}

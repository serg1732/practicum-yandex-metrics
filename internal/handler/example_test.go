package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
)

type exampleReadStorage struct {
	counters map[string]*models.Metrics
	gauges   map[string]*models.Metrics
}

func (s exampleReadStorage) GetCounter(ctx context.Context, name string) (*models.Metrics, error) {
	return s.counters[name], nil
}

func (s exampleReadStorage) GetGauge(ctx context.Context, name string) (*models.Metrics, error) {
	return s.gauges[name], nil
}

func (s exampleReadStorage) GetAllCounters(ctx context.Context) (map[string]*models.Metrics, error) {
	return s.counters, nil
}

func (s exampleReadStorage) GetAllGauges(ctx context.Context) (map[string]*models.Metrics, error) {
	return s.gauges, nil
}

func ExampleBuildReadHandler() {
	var delta int64 = 42
	var value float64 = 12.5

	storage := exampleReadStorage{
		counters: map[string]*models.Metrics{
			"PollCount": {
				ID:    "PollCount",
				MType: models.Counter,
				Delta: &delta,
			},
		},
		gauges: map[string]*models.Metrics{
			"Alloc": {
				ID:    "Alloc",
				MType: models.Gauge,
				Value: &value,
			},
		},
	}

	handler := BuildReadHandler(storage)

	fmt.Printf("%T", handler)

	// Output:
	// handler.ReadMetricsHandlerImpl
}

func ExampleReadMetricsHandlerImpl_SelectMetricHandler_counter() {
	var delta int64 = 42

	storage := exampleReadStorage{
		counters: map[string]*models.Metrics{
			"PollCount": {
				ID:    "PollCount",
				MType: models.Counter,
				Delta: &delta,
			},
		},
	}

	h := BuildReadHandler(storage)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/value/{metricType}/{metricName}", h.SelectMetricHandler(log))

	req := httptest.NewRequest(http.MethodGet, "/value/counter/PollCount", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// 42
}

func ExampleReadMetricsHandlerImpl_SelectMetricHandler_gauge() {
	var value float64 = 12.5

	storage := exampleReadStorage{
		gauges: map[string]*models.Metrics{
			"Alloc": {
				ID:    "Alloc",
				MType: models.Gauge,
				Value: &value,
			},
		},
	}

	h := BuildReadHandler(storage)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/value/{metricType}/{metricName}", h.SelectMetricHandler(log))

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/Alloc", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// 12.5
}

func ExampleReadMetricsHandlerImpl_SelectValueMetricHandler_counter() {
	var delta int64 = 42

	storage := exampleReadStorage{
		counters: map[string]*models.Metrics{
			"PollCount": {
				ID:    "PollCount",
				MType: models.Counter,
				Delta: &delta,
			},
		},
	}

	h := BuildReadHandler(storage)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	body := bytes.NewBufferString(`{"id":"PollCount","type":"counter"}`)
	req := httptest.NewRequest(http.MethodPost, "/value/", body)
	rec := httptest.NewRecorder()

	h.SelectValueMetricHandler(log).ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// {"id":"PollCount","type":"counter","delta":42}
}

func ExampleReadMetricsHandlerImpl_SelectValueMetricHandler_gauge() {
	var value float64 = 12.5

	storage := exampleReadStorage{
		gauges: map[string]*models.Metrics{
			"Alloc": {
				ID:    "Alloc",
				MType: models.Gauge,
				Value: &value,
			},
		},
	}

	h := BuildReadHandler(storage)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	body := bytes.NewBufferString(`{"id":"Alloc","type":"gauge"}`)
	req := httptest.NewRequest(http.MethodPost, "/value/", body)
	rec := httptest.NewRecorder()

	h.SelectValueMetricHandler(log).ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// {"id":"Alloc","type":"gauge","value":12.5}
}

func ExampleReadMetricsHandlerImpl_PingDatabase() {
	h := BuildReadHandler(exampleReadStorage{})
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	h.PingDatabase(log, nil).ServeHTTP(rec, req)

	fmt.Println(rec.Code)

	// Output:
	// 200
}

type exampleUpdateStorage struct {
	metrics []*models.Metrics
}

func (s *exampleUpdateStorage) Update(ctx context.Context, log *slog.Logger, data *models.Metrics) error {
	s.metrics = append(s.metrics, data)
	return nil
}

func (s *exampleUpdateStorage) Updates(ctx context.Context, log *slog.Logger, data []*models.Metrics) error {
	s.metrics = append(s.metrics, data...)
	return nil
}

type exampleAuditor struct {
	events []*models.AuditEvent
}

func (a *exampleAuditor) BroadCast(data *models.AuditEvent) {
	a.events = append(a.events, data)
}

func ExampleBuildUpdateHandler() {
	storage := &exampleUpdateStorage{}
	auditor := &exampleAuditor{}

	handler := BuildUpdateHandler(storage, auditor)

	fmt.Printf("%T", handler)

	// Output:
	// handler.UpdateHandlerImpl
}

func ExampleUpdateHandlerImpl_UpdatePathValuesHandler_gauge() {
	storage := &exampleUpdateStorage{}
	auditor := &exampleAuditor{}
	handler := BuildUpdateHandler(storage, auditor)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12.5", nil)
	req.SetPathValue("metricType", "gauge")
	req.SetPathValue("metricName", "Alloc")
	req.SetPathValue("metricValue", "12.5")

	rec := httptest.NewRecorder()

	handler.UpdatePathValuesHandler(log).ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(storage.metrics[0].ID)
	fmt.Println(storage.metrics[0].MType)
	fmt.Println(*storage.metrics[0].Value)
	fmt.Println(auditor.events[0].Metrics[0])

	// Output:
	// 200
	// Alloc
	// gauge
	// 12.5
	// Alloc
}

func ExampleUpdateHandlerImpl_UpdatePathValuesHandler_counter() {
	storage := &exampleUpdateStorage{}
	auditor := &exampleAuditor{}
	handler := BuildUpdateHandler(storage, auditor)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/42", nil)
	req.SetPathValue("metricType", "counter")
	req.SetPathValue("metricName", "PollCount")
	req.SetPathValue("metricValue", "42")

	rec := httptest.NewRecorder()

	handler.UpdatePathValuesHandler(log).ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(storage.metrics[0].ID)
	fmt.Println(storage.metrics[0].MType)
	fmt.Println(*storage.metrics[0].Delta)
	fmt.Println(auditor.events[0].Metrics[0])

	// Output:
	// 200
	// PollCount
	// counter
	// 42
	// PollCount
}

func ExampleUpdateHandlerImpl_UpdateJSONHandler_gauge() {
	storage := &exampleUpdateStorage{}
	auditor := &exampleAuditor{}
	handler := BuildUpdateHandler(storage, auditor)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	body := bytes.NewBufferString(`{"id":"Alloc","type":"gauge","value":12.5}`)
	req := httptest.NewRequest(http.MethodPost, "/update/", body)
	rec := httptest.NewRecorder()

	handler.UpdateJSONHandler(log).ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(storage.metrics[0].ID)
	fmt.Println(storage.metrics[0].MType)
	fmt.Println(*storage.metrics[0].Value)
	fmt.Println(auditor.events[0].Metrics[0])

	// Output:
	// 200
	// Alloc
	// gauge
	// 12.5
	// Alloc
}

func ExampleUpdateHandlerImpl_UpdateJSONHandler_counter() {
	storage := &exampleUpdateStorage{}
	auditor := &exampleAuditor{}
	handler := BuildUpdateHandler(storage, auditor)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	body := bytes.NewBufferString(`{"id":"PollCount","type":"counter","delta":42}`)
	req := httptest.NewRequest(http.MethodPost, "/update/", body)
	rec := httptest.NewRecorder()

	handler.UpdateJSONHandler(log).ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(storage.metrics[0].ID)
	fmt.Println(storage.metrics[0].MType)
	fmt.Println(*storage.metrics[0].Delta)
	fmt.Println(auditor.events[0].Metrics[0])

	// Output:
	// 200
	// PollCount
	// counter
	// 42
	// PollCount
}

func ExampleUpdateHandlerImpl_UpdateValues() {
	storage := &exampleUpdateStorage{}
	auditor := &exampleAuditor{}
	handler := BuildUpdateHandler(storage, auditor)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	body := bytes.NewBufferString(`[
		{"id":"Alloc","type":"gauge","value":12.5},
		{"id":"PollCount","type":"counter","delta":42}
	]`)

	req := httptest.NewRequest(http.MethodPost, "/updates/", body)
	rec := httptest.NewRecorder()

	handler.UpdateValues(log).ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(len(storage.metrics))
	fmt.Println(storage.metrics[0].ID)
	fmt.Println(storage.metrics[1].ID)
	fmt.Println(auditor.events[0].Metrics[0])
	fmt.Println(auditor.events[0].Metrics[1])

	// Output:
	// 200
	// 2
	// Alloc
	// PollCount
	// Alloc
	// PollCount
}

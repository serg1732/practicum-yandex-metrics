package repository

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-resty/resty/v2"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
)

type UpdaterClient interface {
	ExternalUpdateMetrics(log *slog.Logger, updateCounter int64, metrics map[string]float64) error
	ExternalUpdateJSONMetrics(log *slog.Logger, updateCounter int64, metrics map[string]float64) error
	ExternalBatchUpdateJSONMetrics(log *slog.Logger, updateCounter int64, metrics map[string]float64) error
}

type RestyUpdaterClient struct {
	httpClient *resty.Client
	host       string
}

func BuildRestyUpdaterMetric(host string) UpdaterClient {
	return RestyUpdaterClient{httpClient: resty.New(), host: host}
}

func (r RestyUpdaterClient) ExternalUpdateMetrics(log *slog.Logger, updateCounter int64, metrics map[string]float64) error {
	for k, v := range metrics {
		urlModified := fmt.Sprintf("%s/update/%s/%s/%v", r.host, models.Gauge, k, v)
		resp, err := r.httpClient.R().SetHeader("Content-Type", "text/plain").Post(urlModified)
		if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
			log.Debug("Ошибка обновления метрик gauge",
				slog.String("name", k),
				slog.Float64("value", v),
				slog.Any("error", err))
			return errors.New("ошибка отправки метрики gauge")
		}
	}

	resp, err := r.httpClient.R().SetHeader("Content-Type", "text/plain").Post(
		fmt.Sprintf("%s/update/%s/%s/%v", r.host, models.Counter, "PollCount", updateCounter))
	if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
		log.Debug("Ошибка обновления метрик gauge",
			slog.String("name", "PollCount"),
			slog.Int64("value", updateCounter),
			slog.Any("error", err))
		return errors.New("ошибка отправки метрики counter")
	}
	return nil
}

func (r RestyUpdaterClient) ExternalBatchUpdateJSONMetrics(log *slog.Logger, updateCounter int64, metrics map[string]float64) error {
	modelMetrics := make([]*models.Metrics, 0, len(metrics)+1)
	for k, v := range metrics {
		modelMetrics = append(modelMetrics, &models.Metrics{ID: k, MType: models.Gauge, Value: &v})
	}
	modelMetrics = append(modelMetrics, &models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &updateCounter})
	jsonMetrics, err := json.Marshal(modelMetrics)
	if err != nil {
		return errors.New("ошибка конвертации метрики gauge в json")
	}
	gzipMetric, err := gzipBody(jsonMetrics)
	if err != nil {
		return errors.New("ошибка сжатия (gzip) метрики")
	}

	urlModified := fmt.Sprintf("%s/updates/", r.host)
	resp, err := r.httpClient.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(gzipMetric).
		Post(urlModified)
	if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
		log.Debug("Ошибка обновления метрик", "metrics", string(jsonMetrics))
		return errors.New("ошибка отправки метрики gauge")
	}

	return nil
}

func (r RestyUpdaterClient) ExternalUpdateJSONMetrics(log *slog.Logger, updateCounter int64, metrics map[string]float64) error {
	urlModified := fmt.Sprintf("%s/update/", r.host)
	for k, v := range metrics {
		metric := models.Metrics{ID: k, MType: models.Gauge, Value: &v}
		jsonMetric, err := json.Marshal(metric)
		if err != nil {
			return errors.New("ошибка конвертации метрики gauge в json")
		}

		gzipMetric, err := gzipBody(jsonMetric)
		if err != nil {
			return errors.New("ошибка сжатия (gzip) метрики")
		}

		resp, err := r.httpClient.R().
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetBody(gzipMetric).
			Post(urlModified)
		if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
			log.Debug("Ошибка обновления метрик gauge",
				slog.String("name", k),
				slog.Float64("value", v),
				slog.Any("error", err))
			return errors.New("ошибка отправки метрики gauge")
		}
	}

	metric := models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &updateCounter}
	jsonMetric, err := json.Marshal(metric)
	if err != nil {
		return errors.New("error convert to json counter metrics")
	}
	gzipMetric, err := gzipBody(jsonMetric)
	if err != nil {
		return errors.New("error compress gzip metric")
	}
	resp, err := r.httpClient.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(gzipMetric).
		Post(urlModified)
	if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
		log.Debug("Ошибка обновления метрик gauge",
			slog.String("name", "PollCount"),
			slog.Int64("value", updateCounter),
			slog.Any("error", err))
		return errors.New("ошибка отправки метрики counter")
	}
	return nil
}

func gzipBody(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(data); err != nil {
		_ = gz.Close()
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

package repository

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-resty/resty/v2"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
)

// AuditMetricsClient HTTP клиент по отправке событий.
type AuditMetricsClient struct {
	httpClient *resty.Client
	host       string
}

// BuildRestyAuditMetrics создание клиента по отправке событий.
func BuildRestyAuditMetrics(host string) *AuditMetricsClient {
	return &AuditMetricsClient{httpClient: resty.New(), host: host}
}

// SendMetrics отправка события метрик.
func (a *AuditMetricsClient) SendMetrics(logger *slog.Logger, event *models.AuditEvent) {
	rawEvent, errEvent := json.Marshal(event)
	if errEvent != nil {
		logger.Error("Ошибка при обработке события", "error", errEvent)
		return
	}
	resp, err := a.httpClient.R().SetHeader("Content-Type", "application/json").SetBody(rawEvent).Post(a.host)
	if err != nil {
		logger.Error("Ошибка при отправке события", "error", err)
		return
	}
	defer resp.RawResponse.Body.Close()
	if resp.StatusCode() != http.StatusOK {
		logger.Error("Ошибка при отравке события, статус != 200", "status", resp.Status())
		return
	}
	logger.Debug("Успешная отправка события")
}

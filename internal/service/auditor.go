package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/serg1732/practicum-yandex-metrics/internal/repository"
)

type Subscriber interface {
	Notify(logger *slog.Logger, data *models.AuditEvent)
}

func BuildAuditor(logger *slog.Logger, subscribers ...Subscriber) Auditor {
	return Auditor{
		logger:      logger,
		subscribers: subscribers,
	}
}

type Auditor struct {
	logger      *slog.Logger
	subscribers []Subscriber
}

func (a *Auditor) BroadCast(data *models.AuditEvent) {
	for _, sub := range a.subscribers {
		go sub.Notify(a.logger, data)
	}
}

func (a *Auditor) Subscribe(subscriber Subscriber) {
	a.subscribers = append(a.subscribers, subscriber)
}

func BuildHttpSubscriber(httpClient *repository.AuditMetricsClient) Subscriber {
	return &HttpSubscriber{
		client: httpClient,
	}
}

func BuildFileSubscriber(filepath string) Subscriber {
	return &FileSubscribe{
		filepath: filepath,
	}
}

type HttpSubscriber struct {
	client *repository.AuditMetricsClient
}

func (h *HttpSubscriber) Notify(logger *slog.Logger, data *models.AuditEvent) {
	h.client.SendMetrics(logger, data)
}

type FileSubscribe struct {
	filepath string
	sync.Mutex
}

func (f *FileSubscribe) Notify(logger *slog.Logger, data *models.AuditEvent) {
	f.Lock()
	defer f.Unlock()
	file, err := os.OpenFile(f.filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(*data)
	if err != nil {
		logger.Error("Не удалось записать в файл событие", "error", err)
		return
	}
	logger.Debug("Успешная запись в файл")
}

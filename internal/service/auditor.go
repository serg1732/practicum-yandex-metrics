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

// Subscriber представляет интерфейс, отражающий реализацию подписчика(обработчика лога аудита).
type Subscriber interface {
	// Notify функция уведомления об успешной обработке метрики.
	Notify(logger *slog.Logger, data *models.AuditEvent)
}

// BuildAuditor функция создания аудита запросов с паттерном "Наблюдатель".
func BuildAuditor(logger *slog.Logger, subscribers ...Subscriber) Auditor {
	return Auditor{
		logger:      logger,
		subscribers: subscribers,
	}
}

// Auditor аудитор запросов.
type Auditor struct {
	// logger - логгер
	logger *slog.Logger
	// subscribers - набор подписчиков, которых будет уведомлять.
	subscribers []Subscriber
}

// BroadCast функция запуска уведомления всех подписчиков.
func (a *Auditor) BroadCast(data *models.AuditEvent) {
	for _, sub := range a.subscribers {
		go sub.Notify(a.logger, data)
	}
}

// Subscribe функция добавления подписчика.
func (a *Auditor) Subscribe(subscriber Subscriber) {
	a.subscribers = append(a.subscribers, subscriber)
}

// BuildHttpSubscriber создание подписчика, который обрабатывает лог через http запрос.
func BuildHttpSubscriber(httpClient *repository.AuditMetricsClient) Subscriber {
	return &HttpSubscriber{
		client: httpClient,
	}
}

// BuildFileSubscriber создание подписчика, который обрабатывает лог записью в файл.
func BuildFileSubscriber(filepath string) Subscriber {
	return &FileSubscribe{
		filepath: filepath,
	}
}

// HttpSubscriber подписчик с обработкой события по http.
type HttpSubscriber struct {
	client *repository.AuditMetricsClient
}

// Notify обработка события по http.
func (h *HttpSubscriber) Notify(logger *slog.Logger, data *models.AuditEvent) {
	h.client.SendMetrics(logger, data)
}

// FileSubscribe подписчик с обработкой события через файл.
type FileSubscribe struct {
	// filepath путь до файла.
	filepath string
	sync.Mutex
}

// Notify обработка события через файл.
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

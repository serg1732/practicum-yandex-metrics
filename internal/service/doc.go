/*
	Пакет содержит в себе реализации сборщика метрик и сервис аудита.

# Collector

Собирает метрики системы и отправляет их на сервер.
Пример создания сборщика:

	agent := service.BuildCollector()

Описание функции запуска:

	Run(ctx context.Context, log *slog.Logger, agentConfig config.AgentConfig) error

# Auditor

Уведомляет подписчиков об успешной обработке метрики. На данный момент реализованы подписчики файловый и http.
Пример создания аудитора без подписчиков при старте или же с ними:

	audit := service.BuildAuditor(my-logger-slog) // без
	audit := service.BuildAuditor(my-logger-slog, my-subscribes) // с подписчиками

В процессе работы можно добавить новый обработчик:

	audit.Subscribe(my-subscriber)

Метод уведомления подписчиков:

	auditor.BroadCast(&models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   []string{metricName},
			IPAddress: strings.Split(r.RemoteAddr, ":")[0],
		})

Реализованные подписчики файловый и http.
Создание http обработчика:

	subscriber:= BuildHttpSubscriber(repository.BuildRestyAuditMetrics(serverConfig.AuditUrl))

Создание файлового обработчика:

	subscriber:= BuildFileSubscriber(serverConfig.AuditFile)
*/
package service

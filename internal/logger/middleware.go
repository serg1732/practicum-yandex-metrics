package logger

import (
	"log/slog"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func WithLogger() func(http.Handler) http.Handler {
	logger := slog.Default()
	return func(h http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			logger.With(
				slog.String("URI", r.RequestURI),
				slog.String("method", r.Method),
				slog.Duration("duration", duration),
			).Info("Получен запрос")

			logger.With(
				slog.Int("status", responseData.status),
				slog.Int("size", responseData.size),
			).Info("Ответ на запрос")
		}
		return http.HandlerFunc(logFn)
	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
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

func WithLogger(log *slog.Logger) func(http.Handler) http.Handler {
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
			requestID := uuid.New().String()
			log.With(
				slog.String("request_id", requestID),
				slog.String("URI", r.RequestURI),
				slog.String("method", r.Method),
			).Info("Получен запрос")

			log.With(
				slog.String("request_id", requestID),
				slog.Int("status", responseData.status),
				slog.Int("size", responseData.size),
				slog.Duration("duration", duration),
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
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

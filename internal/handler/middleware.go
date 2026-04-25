package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/serg1732/practicum-yandex-metrics/internal/helpers/compress"
)

// WithGzipCompress middleware обработчик для сжатия / получение исходных данных.
func WithGzipCompress(log *slog.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		compressFunc := func(w http.ResponseWriter, r *http.Request) {
			ow := w
			if strings.Contains(r.Header.Get("Accept"), "application/json") ||
				strings.Contains(r.Header.Get("Accept"), "text/html") {
				acceptEncoding := r.Header.Get("Accept-Encoding")
				supportsGzip := strings.Contains(acceptEncoding, "gzip")
				if supportsGzip {
					w.Header().Set("Content-Encoding", "gzip")
					cw := compress.NewCompressWriter(w)
					ow = cw
					defer cw.Close()
				}

				contentEncoding := r.Header.Get("Content-Encoding")
				sendsGzip := strings.Contains(contentEncoding, "gzip")
				if sendsGzip {
					cr, err := compress.NewCompressReader(r.Body)
					if err != nil {
						log.Error("Ошибка при создании декомпрессора", "error", err.Error())
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					r.Body = cr
					defer cr.Close()
				}
			}
			h.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(compressFunc)
	}
}

type ResponseWithBody struct {
	http.ResponseWriter
	body []byte
}

func (r *ResponseWithBody) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.body = append(r.body, b...)
	return size, err
}

// WithCheckHash middleware по проверки hash значения запроса.
func WithCheckHash(log *slog.Logger, key string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Error("Ошибка чтения body", "error", err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			hash, errHash := getHash(body, key)
			if errHash != nil {
				log.Error("Ошибка при получении hash", "error", errHash.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			requestHash := r.Header.Get("HashSHA256")

			if key != "" && requestHash != "" && requestHash != hash {
				log.Debug("Не совпадают hash значения")
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			cw := ResponseWithBody{
				ResponseWriter: w,
				body:           make([]byte, 0),
			}
			h.ServeHTTP(&cw, r)

			if key != "" {
				responseHash, errResponseHash := getHash(cw.body, key)
				if errResponseHash != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
				w.Header().Set("HashSHA256", responseHash)
			}
		})
	}
}

// getHash вспомогательная функция получения hash значения.
func getHash(data []byte, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))

	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	hash := h.Sum(nil)

	return hex.EncodeToString(hash), nil
}

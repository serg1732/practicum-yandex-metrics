package handler

import (
	"net/http"
	"strings"

	"github.com/serg1732/practicum-yandex-metrics/internal/helpers/compress"
)

func WithGzipCompress() func(http.Handler) http.Handler {
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

package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

var gzipPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
}

type compressWriter struct {
	w           http.ResponseWriter
	zw          *gzip.Writer
	wroteHeader bool
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true

	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.Header().Del("Content-Length")

	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) initWriter() {
	if c.zw != nil {
		return
	}

	zw := gzipPool.Get().(*gzip.Writer)
	zw.Reset(c.w)
	c.zw = zw
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}

	c.initWriter()
	return c.zw.Write(p)
}

func (c *compressWriter) Close() error {
	if c.zw == nil {
		return nil
	}

	err := c.zw.Close()

	c.zw.Reset(io.Discard)
	gzipPool.Put(c.zw)

	return err
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

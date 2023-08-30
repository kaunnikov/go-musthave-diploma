package compression

import (
	"compress/gzip"
	"io"
	"kaunnikov/internal/logging"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

var successCompressionContentType = [2]string{"application/json", "text/html"}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func CustomCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isGoodContentType := false
		contentType := r.Header.Get("Content-Type")

		// Проверям, ожидает ли клиент, что сервер будет сжимать данные gzip
		isNeedCompression := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		// Проверяем тип контента
		for _, c := range successCompressionContentType {
			if contentType == c {
				isGoodContentType = true
				break
			}
		}

		// Проверяем нужно ли раскодировать данные, которые прислал клиент
		if r.Header.Get(`Content-Encoding`) == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				logging.Errorf("Error NewReader(body): %s", err)
				return
			}
			r.Body = gz
			defer func(gz *gzip.Reader) {
				err := gz.Close()
				if err != nil {
					logging.Errorf("Error gz.Close: %s", err)
				}
			}(gz)
		}

		// Если условия для сжатия не выполнены - отдаём ответ
		if !isNeedCompression || !isGoodContentType {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			logging.Errorf("Error gzip compression: %s", err)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				logging.Errorf("Error gz close: %s", err)
			}
		}(gz)

		h.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

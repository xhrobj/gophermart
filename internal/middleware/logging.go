package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	if w.responseData.status == 0 {
		w.responseData.status = http.StatusOK
	}

	size, err := w.ResponseWriter.Write(data)
	w.responseData.size += size

	return size, err
}

// WithLogging возвращает middleware для логирования HTTP-запросов.
func WithLogging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()

			responseData := &responseData{
				status: http.StatusOK,
			}

			loggingWriter := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			next.ServeHTTP(loggingWriter, r)

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
				zap.Duration("duration", time.Since(startedAt)),
			)
		})
	}
}

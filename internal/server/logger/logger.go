package logger

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseData stores information about HTTP response
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter is a wrapper over http.ResponseWriter to track the response
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write writes data and keeps track of its size
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader sets the status of the response and writes it down
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging returns an HTTP handler that adds logging
func WithLogging(sugar *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			sugar.Infow("HTTP request info",
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
			)
		})
	}
}

// getSyncFunc возвращает функцию для синхронизации и закрытия логгера.
func getSyncFunc(logger *zap.Logger) func() {
	return func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Can't sync log: %v", err)
		}
	}
}

// initLogger инициализирует и возвращает SugaredLogger и функцию для его синхронизации.
func InitLogger() (*zap.SugaredLogger, func(), error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, err
	}
	sugar := zapLogger.Sugar()
	syncFunc := getSyncFunc(zapLogger)
	return sugar, syncFunc, nil
}

package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/constants"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

			var formattedBody []byte
			if r.Header.Get("Content-Type") == constants.ApplicationJSON {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					sugar.Errorw("Cannot read request body", "err", err)
				}

				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				var raw json.RawMessage
				if err = json.Unmarshal(bodyBytes, &raw); err != nil {
					sugar.Errorw("Cannot unmarshal request body", "err", err)
				}

				var singleMetric models.Metrics
				var sliceMetrics []models.Metrics

				if err = json.Unmarshal(raw, &singleMetric); err == nil {
					formattedBody, err = json.Marshal(singleMetric)
				} else if err = json.Unmarshal(raw, &sliceMetrics); err == nil {
					formattedBody, err = json.Marshal(sliceMetrics)
				} else {
					sugar.Errorw("Unknown type of metrics", "type", fmt.Sprintf("%T", raw))
				}

				if err != nil {
					sugar.Errorw("Cannot marshal request body", "err", err)
				}
			}

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
				"timestamp", start.Format("2006-01-02 15:04:05"),
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
				"body", string(formattedBody),
			)
		})
	}
}

// getSyncFunc returns a function to synchronize and close the logger
func getSyncFunc(logger *zap.Logger) func() {
	return func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Can't sync log: %v", err)
		}
	}
}

// initLogger initializes and returns SugaredLogger and a function to synchronize it
func InitLogger(cfg *config.Config) (*zap.SugaredLogger, func(), error) {
	var zapLogger *zap.Logger
	var err error

	if cfg.Environment == "production" {
		zapLogger, err = zap.NewProduction()

	} else {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = utils.CustomTimeEncoder
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder

		encoder := zapcore.NewConsoleEncoder(encoderConfig)

		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(zapcore.DebugLevel),
		)

		zapLogger = zap.New(core)
	}

	if err != nil {
		return nil, nil, err
	}

	sugar := zapLogger.Sugar()
	syncFunc := getSyncFunc(zapLogger)
	return sugar, syncFunc, nil
}

package retry

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"go.uber.org/zap"
)

type Error struct {
	LastErr  error
	ErrorMsg string
	Retries  int
}

func (e *Error) Error() string {
	return fmt.Sprintf("After %d retries: %s. Last error: %s", e.Retries, e.ErrorMsg, e.LastErr)
}

// Operation performs a retryable operation based on the given config
func Operation(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger, operation func() error) error {
	var err error
	delay := cfg.InitialDelay * time.Second
	increment := cfg.DelayIncrement * time.Second
	retries := 0

	for i := 0; i < cfg.MaxRetries; i++ {
		retries++
		sugar.Infof("Attempting operation, attempt %d", retries)

		err = operation()
		if err == nil {
			sugar.Info("Operation successful")
			return nil
		}
		if !shouldRetry(err) {
			sugar.Warnf("Not retrying due to error: %s", err)
			return err
		}

		sugar.Warnf("Operation failed: %s. Retrying...", err)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			sugar.Warn("Operation cancelled")
			return ctx.Err()
		}

		delay += increment
	}

	sugar.Error("Operation ultimately failed after max retries")
	return &Error{
		LastErr:  err,
		Retries:  retries,
		ErrorMsg: "Operation ultimately failed",
	}
}

func shouldRetry(err error) bool {
	_, ok1 := err.(net.Error)

	return ok1
}

// WithRetry returns a new retry middleware with given retry config
func WithRetry(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			operation := func() error {
				rec := httptest.NewRecorder()
				next.ServeHTTP(rec, r)
				for k, v := range rec.Header() {
					w.Header()[k] = v
				}
				w.WriteHeader(rec.Code)
				_, err := w.Write(rec.Body.Bytes())
				return err
			}

			if err := Operation(ctx, cfg, sugar, operation); err != nil {
				sugar.Errorf("Internal Server Error: %s", err, "WithRetry")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		})
	}
}

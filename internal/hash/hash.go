package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"go.uber.org/zap"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	key          string
	hashInHeader bool
}

func (rwi *responseWriterInterceptor) Write(p []byte) (int, error) {
	if rwi.hashInHeader {
		rwi.addHashSHA256Header(p)
	}

	return rwi.ResponseWriter.Write(p)
}

func (rwi *responseWriterInterceptor) addHashSHA256Header(p []byte) {
	responseHash := ComputeHash(p, rwi.key)
	rwi.Header().Set("HashSHA256", responseHash)
}

func ComputeHash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

func WithHashing(cfg *config.Config, sugar *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hashInHeader := r.Header.Get("HashSHA256") != ""

			if hashInHeader {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					sugar.Errorw("failed to read request body", "error", err)
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				computedHash := ComputeHash(body, cfg.Key)
				sugar.Infof("received hash: %s, computed hash: %s", r.Header.Get("HashSHA256"), computedHash)

				if r.Header.Get("HashSHA256") != computedHash {
					sugar.Warnw("invalid hash", "received", r.Header.Get("HashSHA256"), "computed", computedHash)
					http.Error(w, "invalid hash", http.StatusBadRequest)
					return
				}
				sugar.Infow("hash validated", "hash", computedHash)
			}

			rwi := &responseWriterInterceptor{ResponseWriter: w, hashInHeader: hashInHeader, key: cfg.Key}
			next.ServeHTTP(rwi, r)
		})
	}
}

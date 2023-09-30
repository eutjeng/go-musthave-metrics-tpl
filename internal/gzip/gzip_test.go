package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestWithCompression(t *testing.T) {
	sugar := zap.NewExample().Sugar()
	middleware := WithCompression(sugar)

	tests := []struct {
		name                  string
		acceptEncoding        string
		expectContentEncoding string
		body                  string
	}{
		{"With Gzip", "gzip", "gzip", "Hello, world!"},
		{"Without Gzip", "", "", "Hello, world!"},
		{"With Gzip and Big Data", "gzip", "gzip", strings.Repeat("Hello, world! ", 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.body))
			})

			middleware(handler).ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			if got := res.Header.Get("Content-Encoding"); got != tt.expectContentEncoding {
				t.Errorf("Content-Encoding = %v, want %v", got, tt.expectContentEncoding)
			}

			got, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Could not read response: %v", err)
			}

			if tt.expectContentEncoding == "gzip" {
				gr, err := gzip.NewReader(bytes.NewBuffer(got))
				if err != nil {
					t.Fatalf("Could not create gzip reader: %v", err)
				}
				got, err = io.ReadAll(gr)
				if err != nil {
					t.Fatalf("Could not read gzipped response: %v", err)
				}
			}

			if string(got) != tt.body {
				t.Errorf("Unexpected response body: got %v want %v", string(got), tt.body)
			}

			originalSize := len(tt.body)
			compressedSize := rr.Body.Len()

			if tt.expectContentEncoding == "gzip" {
				if originalSize > 50 && compressedSize >= originalSize {
					t.Errorf("Compressed response is not smaller: got %v want smaller than %v", compressedSize, originalSize)
				}
			} else if compressedSize != originalSize {
				t.Errorf("Response size should be the same as original when not compressed: got %v want %v", compressedSize, originalSize)
			}

		})
	}
}

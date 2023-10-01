package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// compressWriter implements http.ResponseWriter and provides transparent
// compression for server responses, setting appropriate HTTP headers
type compressWriter struct {
	w          http.ResponseWriter // the original http.ResponseWriter
	zw         *gzip.Writer        // gzip writer to compress the data
	statusCode int                 // HTTP status code to set when writing the header
}

// compressReader implements io.ReadCloser and provides transparent
// decompression for incoming client data
type compressReader struct {
	r  io.ReadCloser // the original ReadCloser
	zr *gzip.Reader  // gzip reader to decompress the data
}

// newCompressWriter creates a new compressWriter given an http.ResponseWriter
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the headers from the original http.ResponseWriter
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes the data, compressing it if necessary
func (c *compressWriter) Write(p []byte) (int, error) {
	if c.statusCode >= http.StatusOK && c.statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.WriteHeader(c.statusCode)
		return c.zw.Write(p)
	}

	c.w.WriteHeader(c.statusCode)
	return c.w.Write(p)
}

// WriteHeader sets the HTTP status code for the response
func (c *compressWriter) WriteHeader(statusCode int) {
	c.statusCode = statusCode
}

// Close closes the gzip.Writer and flushes any buffered data
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// newCompressReader creates a new compressReader given an io.ReadCloser
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads from the decompressed data stream
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the original ReadCloser and the gzip.Reader
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// WithCompression returns a middleware that adds gzip compression and
// decompression to the request/response cycle
func WithCompression(sugar *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			var cw *compressWriter
			if supportsGzip {
				cw = newCompressWriter(w)
				ow = cw
				defer func() {
					if cw.statusCode >= http.StatusOK && cw.statusCode < http.StatusMultipleChoices {
						if err := cw.Close(); err != nil {
							sugar.Errorw("Failed to close compress writer", err)
							return
						}
						cw.zw.Reset(cw.w)
					}
				}()
				sugar.Infow("Compression applied", "encoding", "gzip")
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					sugar.Errorw("Failed to decompress request", err)
					return
				}
				r.Body = cr
				defer cr.Close()
				sugar.Infow("Decompression applied", "encoding", "gzip")
			}

			next.ServeHTTP(ow, r)
		})
	}
}

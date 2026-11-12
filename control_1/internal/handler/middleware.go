package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"
)

type requestIDKey struct{}

var requestCounter atomic.Uint64

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}

	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(data)
}

func newRequestID() string {
	var bytes [16]byte

	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}

	return fmt.Sprintf(
		"%d-%d",
		time.Now().UnixNano(),
		requestCounter.Add(1),
	)
}

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := newRequestID()

			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
			r = r.WithContext(ctx)

			w.Header().Set("X-Request-ID", requestID)

			recorder := &responseRecorder{
				ResponseWriter: w,
			}

			defer func() {
				if recorder.status == 0 {
					recorder.status = http.StatusOK
				}

				logger.Info(
					"HTTP request",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", recorder.status),
					slog.Duration("duration", time.Since(start)),
					slog.String("request_id", requestID),
				)
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					requestID, _ := r.Context().Value(requestIDKey{}).(string)

					logger.Error(
						"HTTP handler panic",
						slog.Any("panic", recovered),
						slog.String("request_id", requestID),
						slog.String("stack", string(debug.Stack())),
					)

					writeError(
						w,
						http.StatusInternalServerError,
						"INTERNAL_ERROR",
						"internal server error",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

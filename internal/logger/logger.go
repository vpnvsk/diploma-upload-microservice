package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func LoggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		log.Info(fmt.Sprintf("%q %d %s",
			fmt.Sprintf("%s %s %s", r.Method, r.URL.Path, r.Proto),
			lrw.statusCode,
			time.Since(start),
		))
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

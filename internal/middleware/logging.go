package middleware

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

type errorPage struct {
	StatusCode    int
	StatusMessage string
}

func HandleBadCode(tmpl *template.Template, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		if wrapped.statusCode >= 400 {
			tmpl.ExecuteTemplate(w, "error.html", errorPage{
				StatusCode:    wrapped.statusCode,
				StatusMessage: http.StatusText(wrapped.statusCode),
			})
		}
	})
}

func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		logger.Info(
			"handled request",
			slog.Int("statusCode", wrapped.statusCode),
			slog.String("remoteAddr", r.RemoteAddr),
			slog.String("xffHeader", r.Header.Get("X-Forwarded-For")),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("duration", time.Since(start)),
		)
	})
}

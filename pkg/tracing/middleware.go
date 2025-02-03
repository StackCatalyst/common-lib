package tracing

import (
	"fmt"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// HTTPMiddleware creates middleware for tracing HTTP requests
func HTTPMiddleware(tracer *Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to extract span context from headers
			spanCtx, err := tracer.Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(r.Header),
			)

			// Start the span
			var span opentracing.Span
			if err != nil {
				// No span context in headers, create a new one
				span = tracer.StartSpan(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			} else {
				// Create span as child of extracted context
				span = tracer.StartSpan(
					fmt.Sprintf("%s %s", r.Method, r.URL.Path),
					ext.RPCServerOption(spanCtx),
				)
			}
			defer span.Finish()

			// Set standard HTTP tags
			ext.HTTPMethod.Set(span, r.Method)
			ext.HTTPUrl.Set(span, r.URL.String())
			ext.Component.Set(span, "http")

			// Add request-specific tags
			span.SetTag("http.remote_addr", r.RemoteAddr)
			span.SetTag("http.user_agent", r.UserAgent())
			if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
				span.SetTag("http.request_id", reqID)
			}

			// Create wrapped response writer to capture status code
			wrapped := wrapResponseWriter(w)

			// Add span to request context
			ctx := opentracing.ContextWithSpan(r.Context(), span)
			r = r.WithContext(ctx)

			// Record timing
			start := time.Now()

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Add response tags
			duration := time.Since(start)
			span.SetTag("http.status_code", wrapped.status)
			span.SetTag("http.duration_ms", float64(duration.Milliseconds()))

			// Mark error if status >= 500
			if wrapped.status >= http.StatusInternalServerError {
				ext.Error.Set(span, true)
				span.SetTag("error.type", "server_error")
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

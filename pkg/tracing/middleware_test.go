package tracing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		requestID     string
		handlerStatus int
		handlerError  error
		expectError   bool
		expectedTags  map[string]interface{}
		parentSpanCtx opentracing.SpanContext
	}{
		{
			name:          "successful request",
			method:        "GET",
			path:          "/test",
			requestID:     "req-123",
			handlerStatus: http.StatusOK,
			expectedTags: map[string]interface{}{
				"http.method":      "GET",
				"http.url":         "/test",
				"http.request_id":  "req-123",
				"http.status_code": http.StatusOK,
				"component":        "http",
			},
		},
		{
			name:          "server error",
			method:        "POST",
			path:          "/error",
			handlerStatus: http.StatusInternalServerError,
			handlerError:  fmt.Errorf("internal error"),
			expectError:   true,
			expectedTags: map[string]interface{}{
				"http.method":      "POST",
				"http.url":         "/error",
				"http.status_code": http.StatusInternalServerError,
				"error":            true,
				"error.type":       "server_error",
				"component":        "http",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock tracer
			mockTracer := mocktracer.New()
			tracer := &Tracer{
				tracer: mockTracer,
				config: DefaultConfig(),
			}

			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify span is in context
				span := opentracing.SpanFromContext(r.Context())
				assert.NotNil(t, span)

				if tt.handlerError != nil {
					http.Error(w, tt.handlerError.Error(), tt.handlerStatus)
					return
				}
				w.WriteHeader(tt.handlerStatus)
			})

			// Create test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.requestID != "" {
				req.Header.Set("X-Request-ID", tt.requestID)
			}

			// Add parent span context if provided
			if tt.parentSpanCtx != nil {
				carrier := opentracing.HTTPHeadersCarrier(req.Header)
				err := tracer.tracer.Inject(tt.parentSpanCtx, opentracing.HTTPHeaders, carrier)
				require.NoError(t, err)
			}

			// Create response recorder
			rec := httptest.NewRecorder()

			// Apply middleware
			middleware := HTTPMiddleware(tracer)
			middleware(handler).ServeHTTP(rec, req)

			// Verify response
			assert.Equal(t, tt.handlerStatus, rec.Code)

			// Check spans
			spans := mockTracer.FinishedSpans()
			require.Len(t, spans, 1)
			span := spans[0]

			// Verify operation name
			expectedOp := fmt.Sprintf("%s %s", tt.method, tt.path)
			assert.Equal(t, expectedOp, span.OperationName)

			// Verify tags
			tags := span.Tags()
			for k, v := range tt.expectedTags {
				assert.Equal(t, v, tags[k], "tag %s mismatch", k)
			}

			// Verify timing
			assert.Greater(t, tags["http.duration_ms"], float64(0))

			// Verify parent span context
			if tt.parentSpanCtx != nil {
				mockCtx, ok := tt.parentSpanCtx.(*mocktracer.SpanContext)
				require.True(t, ok)
				assert.Equal(t, mockCtx.SpanID, span.ParentID)
			}
		})
	}
}

func TestResponseWriter(t *testing.T) {
	tests := []struct {
		name         string
		writeHeader  bool
		writeBody    bool
		statusCode   int
		body         string
		expectStatus int
	}{
		{
			name:         "explicit status no body",
			writeHeader:  true,
			statusCode:   http.StatusCreated,
			expectStatus: http.StatusCreated,
		},
		{
			name:         "body without status",
			writeBody:    true,
			body:         "test body",
			expectStatus: http.StatusOK,
		},
		{
			name:         "status and body",
			writeHeader:  true,
			writeBody:    true,
			statusCode:   http.StatusAccepted,
			body:         "test body",
			expectStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			wrapped := wrapResponseWriter(rec)

			if tt.writeHeader {
				wrapped.WriteHeader(tt.statusCode)
			}
			if tt.writeBody {
				_, err := wrapped.Write([]byte(tt.body))
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectStatus, wrapped.Status())
			if tt.writeBody {
				assert.Equal(t, tt.body, rec.Body.String())
			}
		})
	}
}

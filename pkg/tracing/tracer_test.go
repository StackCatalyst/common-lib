package tracing

import (
	"context"
	"fmt"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracerCreation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "default config",
			config: Config{
				ServiceName:  "test-service",
				AgentHost:    "localhost",
				AgentPort:    "6831",
				Enabled:      true,
				SamplingRate: 0.1,
			},
			wantErr: false,
		},
		{
			name: "disabled tracer",
			config: Config{
				Enabled: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tracer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tracer)
				assert.NoError(t, tracer.Close())
			}
		})
	}
}

func TestSpanCreation(t *testing.T) {
	// Use mock tracer for testing
	mockTracer := mocktracer.New()
	tracer := &Tracer{
		tracer: mockTracer,
		config: DefaultConfig(),
	}

	// Test simple span creation
	span := tracer.StartSpan("test-operation")
	require.NotNil(t, span)
	span.Finish()

	mockSpans := mockTracer.FinishedSpans()
	require.Len(t, mockSpans, 1)
	assert.Equal(t, "test-operation", mockSpans[0].OperationName)
}

func TestSpanContext(t *testing.T) {
	mockTracer := mocktracer.New()
	tracer := &Tracer{
		tracer: mockTracer,
		config: DefaultConfig(),
	}

	// Create parent span
	parentSpan, parentCtx := tracer.StartSpanFromContext(context.Background(), "parent-operation")
	require.NotNil(t, parentSpan)

	// Create child span
	childSpan, _ := tracer.StartSpanFromContext(parentCtx, "child-operation")
	require.NotNil(t, childSpan)

	// Finish spans
	childSpan.Finish()
	parentSpan.Finish()

	// Verify spans
	mockSpans := mockTracer.FinishedSpans()
	require.Len(t, mockSpans, 2)
	assert.Equal(t, "child-operation", mockSpans[0].OperationName)
	assert.Equal(t, "parent-operation", mockSpans[1].OperationName)
}

func TestSpanTags(t *testing.T) {
	mockTracer := mocktracer.New()
	tracer := &Tracer{
		tracer: mockTracer,
		config: DefaultConfig(),
	}

	// Create context with span
	span, ctx := tracer.StartSpanFromContext(context.Background(), "test-operation")
	require.NotNil(t, span)

	// Add fields
	WithField(ctx, "string-tag", "value")
	WithField(ctx, "int-tag", 42)
	WithError(ctx, fmt.Errorf("test error"))
	WithFields(ctx, map[string]interface{}{
		"batch-tag-1": "value1",
		"batch-tag-2": "value2",
	})

	span.Finish()

	// Verify tags
	mockSpans := mockTracer.FinishedSpans()
	require.Len(t, mockSpans, 1)
	mockSpan := mockSpans[0]

	assert.Equal(t, "value", mockSpan.Tags()["string-tag"])
	assert.Equal(t, 42, mockSpan.Tags()["int-tag"])
	assert.Equal(t, true, mockSpan.Tags()["error"])
	assert.Equal(t, "test error", mockSpan.Tags()["error.message"])
	assert.Equal(t, "value1", mockSpan.Tags()["batch-tag-1"])
	assert.Equal(t, "value2", mockSpan.Tags()["batch-tag-2"])
}

func TestContextPropagation(t *testing.T) {
	mockTracer := mocktracer.New()
	tracer := &Tracer{
		tracer: mockTracer,
		config: DefaultConfig(),
	}

	// Create a span and inject it into carrier
	span := tracer.StartSpan("test-operation")
	carrier := opentracing.TextMapCarrier{}
	err := tracer.tracer.Inject(span.Context(), opentracing.TextMap, carrier)
	require.NoError(t, err)

	// Extract span context from carrier
	spanContext, err := tracer.Extract(opentracing.TextMap, carrier)
	require.NoError(t, err)

	// Create a new span with extracted context
	childSpan := tracer.StartSpan("child-operation",
		opentracing.ChildOf(spanContext))
	require.NotNil(t, childSpan)

	// Finish spans
	childSpan.Finish()
	span.Finish()

	// Verify spans
	mockSpans := mockTracer.FinishedSpans()
	require.Len(t, mockSpans, 2)
	assert.Equal(t, mockSpans[1].SpanContext.SpanID, mockSpans[0].ParentID)
}

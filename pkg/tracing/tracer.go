package tracing

import (
	"context"
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
)

// Config holds the configuration for the tracer
type Config struct {
	// ServiceName is the name of the service
	ServiceName string
	// AgentHost is the host of the Jaeger agent
	AgentHost string
	// AgentPort is the port of the Jaeger agent
	AgentPort string
	// Enabled determines if tracing is enabled
	Enabled bool
	// SamplingRate is the rate at which traces are sampled (0.0 to 1.0)
	SamplingRate float64
	// Tags are additional tags to add to all spans
	Tags map[string]string
}

// DefaultConfig returns the default tracer configuration
func DefaultConfig() Config {
	return Config{
		ServiceName:  "service",
		AgentHost:    "localhost",
		AgentPort:    "6831",
		Enabled:      true,
		SamplingRate: 0.1,
		Tags:         make(map[string]string),
	}
}

// Tracer manages the distributed tracing functionality
type Tracer struct {
	tracer opentracing.Tracer
	closer io.Closer
	config Config
}

// New creates a new tracer
func New(cfg Config) (*Tracer, error) {
	if !cfg.Enabled {
		return &Tracer{
			tracer: opentracing.NoopTracer{},
			config: cfg,
		}, nil
	}

	jcfg := &config.Configuration{
		ServiceName: cfg.ServiceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: cfg.SamplingRate,
		},
		Reporter: &config.ReporterConfig{
			LocalAgentHostPort: fmt.Sprintf("%s:%s", cfg.AgentHost, cfg.AgentPort),
		},
		Tags: []opentracing.Tag{
			{Key: "service.version", Value: "1.0.0"},
			{Key: "environment", Value: "production"},
		},
	}

	// Add custom tags
	for k, v := range cfg.Tags {
		jcfg.Tags = append(jcfg.Tags, opentracing.Tag{Key: k, Value: v})
	}

	// Initialize metrics factory
	metricsFactory := prometheus.New()

	// Initialize tracer
	tracer, closer, err := jcfg.NewTracer(
		config.Metrics(metricsFactory),
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}

	// Set as global tracer
	opentracing.SetGlobalTracer(tracer)

	return &Tracer{
		tracer: tracer,
		closer: closer,
		config: cfg,
	}, nil
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(name string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return t.tracer.StartSpan(name, opts...)
}

// StartSpanFromContext starts a new span from a context
func (t *Tracer) StartSpanFromContext(ctx context.Context, name string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	}
	span := t.StartSpan(name, opts...)
	return span, opentracing.ContextWithSpan(ctx, span)
}

// Inject injects span context into carrier
func (t *Tracer) Inject(ctx context.Context, format interface{}, carrier interface{}) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return fmt.Errorf("no span in context")
	}
	return t.tracer.Inject(span.Context(), format, carrier)
}

// Extract extracts span context from carrier
func (t *Tracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return t.tracer.Extract(format, carrier)
}

// Close closes the tracer
func (t *Tracer) Close() error {
	if t.closer != nil {
		return t.closer.Close()
	}
	return nil
}

// WithField adds a field to the span in context
func WithField(ctx context.Context, key string, value interface{}) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.SetTag(key, value)
	}
}

// WithError adds an error to the span in context
func WithError(ctx context.Context, err error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())
	}
}

// WithFields adds multiple fields to the span in context
func WithFields(ctx context.Context, fields map[string]interface{}) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		for k, v := range fields {
			span.SetTag(k, v)
		}
	}
}

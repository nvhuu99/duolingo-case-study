package trace

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	traceManager      *TraceManager
	traceManagerReady atomic.Bool
)

func InitTraceManager(
	ctx context.Context,
	resource *resource.Resource,
	traceExporter *otlptrace.Exporter,
) {
	/* New TraceManager */

	if traceManagerReady.Load() {
		return
	}
	defer traceManagerReady.Store(true)

	traceManager = &TraceManager{
		spans:          make(map[string]trace.Span),
		spanDecorators: make(map[SpanNameTemplate][]SpanProcessorFunc),
		spanFinalizers: make(map[SpanNameTemplate][]SpanProcessorFunc),
	}

	/* Setup Otel SDK for Tracing */

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource),
	)
	tracer := traceProvider.Tracer("otlp.grpc.tracer")

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	traceManager.traceProvider = traceProvider
	traceManager.tracer = tracer
}

func GetManager() *TraceManager {
	return traceManager
}

func WithDefaultResource(serviceName string, attrs ...attribute.KeyValue) *resource.Resource {
	resourceAttrs := append(attrs, semconv.ServiceName(serviceName))
	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, resourceAttrs...),
	)
	panicIfErr(err)
	return resource
}

func WithGRPCExporter(endpoint string, secure bool, opts ...otlptracegrpc.Option) *otlptrace.Exporter {
	// Merge exporter options
	if !secure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	mergedOptions := append(opts, otlptracegrpc.WithEndpoint(endpoint))

	// Create exporter
	traceExporter, err := otlptracegrpc.New(context.Background(), mergedOptions...)
	panicIfErr(err)

	return traceExporter
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

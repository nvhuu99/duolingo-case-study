package trace

import (
	"context"

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

type TraceManagerBuilder struct {
	ctx           context.Context
	resource      *resource.Resource
	traceExporter *otlptrace.Exporter
}

func NewTraceManagerBuilder(ctx context.Context) *TraceManagerBuilder {
	return &TraceManagerBuilder{
		ctx: ctx,
	}
}

func (builder *TraceManagerBuilder) WithDefaultResource(
	serviceName string,
	attrs ...attribute.KeyValue,
) *TraceManagerBuilder {
	resourceAttrs := append(attrs, semconv.ServiceName(serviceName))
	resource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, resourceAttrs...),
	)
	panicIfErr(err)
	builder.resource = resource
	return builder
}

func (builder *TraceManagerBuilder) WithGRPCExporter(
	endpoint string,
	secure bool,
	opts ...otlptracegrpc.Option,
) *TraceManagerBuilder {
	if !secure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	mergedOptions := append(opts, otlptracegrpc.WithEndpoint(endpoint))
	traceExporter, err := otlptracegrpc.New(context.Background(), mergedOptions...)
	panicIfErr(err)

	builder.traceExporter = traceExporter

	return builder
}

func (builder *TraceManagerBuilder) GetManager() *TraceManager {
	/* New TraceManager */

	traceManager := &TraceManager{
		spans:          make(map[string]trace.Span),
		spanDecorators: make(map[SpanNameTemplate][]SpanProcessorFunc),
		spanFinalizers: make(map[SpanNameTemplate][]SpanProcessorFunc),
	}

	/* Setup Otel SDK for Tracing */

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(builder.traceExporter),
		sdktrace.WithResource(builder.resource),
	)
	tracer := traceProvider.Tracer("otlp.grpc.tracer")

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	traceManager.traceProvider = traceProvider
	traceManager.tracer = tracer

	return traceManager
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

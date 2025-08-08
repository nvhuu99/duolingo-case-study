package trace

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type SpanDescribeFunc func (span trace.Span, data DataBag)

type TraceManager struct {
	tracer trace.Tracer
	traceProvider *sdktrace.TracerProvider

	spanMutex sync.Mutex
	spans map[string]trace.Span // map by span id
	spanDecribers map[SpanNameTemplate][]SpanDescribeFunc // map by span name template
}

func (m *TraceManager) Shutdown(ctx context.Context) error {
	return m.traceProvider.Shutdown(ctx)
}

func (m *TraceManager) Describe(spanName SpanNameTemplate, describe SpanDescribeFunc) {
	m.spanDecribers[spanName] = append(m.spanDecribers[spanName], describe)
}

func (m *TraceManager) Span(ctx context.Context) trace.Span {
	spanCtx := trace.SpanContextFromContext(ctx)
	return m.spans[spanCtx.SpanID().String()]
}

func (m *TraceManager) Start(
	ctx context.Context, 
	spanName string, 
	timestamp time.Time,
) (context.Context, trace.Span) {
	spanCtx, span := m.tracer.Start(ctx, spanName, trace.WithTimestamp(timestamp))

	m.Track(span)
	
	return spanCtx, span
}

func (m *TraceManager) End(
	span trace.Span, 
	timestamp time.Time,
	statusCode codes.Code,
	message string,
	err error,
	data DataBag,
) {
	span.SetStatus(statusCode, message)
	span.RecordError(err)

	for spanNameTemplate, describers := range m.spanDecribers {
		spanName := span.(sdktrace.ReadOnlySpan).Name()
		if spanNameTemplate.Matches(spanName) {
			for _, describe := range describers {
				describe(span, data.Merge(spanNameTemplate.ExtractVariables(spanName)))
			}
		}
	}

	span.End(trace.WithTimestamp(timestamp))
	
	m.UnTrack(span)
}

func (m *TraceManager) Track(span trace.Span) {
	spanId := span.SpanContext().SpanID().String()
	m.spanMutex.Lock()
	m.spans[spanId] = span
	m.spanMutex.Unlock()
}

func (m *TraceManager) UnTrack(span trace.Span) {
	spanId := span.SpanContext().SpanID().String()
	m.spanMutex.Lock()
	delete(m.spans, spanId)
	m.spanMutex.Unlock()
}


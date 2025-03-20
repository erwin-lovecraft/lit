package instrumentkafka

import (
	"strconv"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// PublishSegment represents a segment of kafka publish operation
type PublishSegment interface {
	SpanID() string

	SpanContext() trace.SpanContext

	End(partition int32, offset int64, sendErr error)
}

type publishSegment struct {
	span   trace.Span
	spanID string
}

func buildPublishSegment(span trace.Span) PublishSegment {
	return publishSegment{
		span:   span,
		spanID: span.SpanContext().SpanID().String(),
	}
}

func (ps publishSegment) SpanID() string {
	return ps.spanID
}

func (ps publishSegment) SpanContext() trace.SpanContext {
	return ps.span.SpanContext()
}

func (ps publishSegment) End(partition int32, offset int64, sendErr error) {
	if sendErr != nil {
		ps.span.RecordError(sendErr, trace.WithStackTrace(true))
	}
	ps.span.SetAttributes(
		semconv.MessagingDestinationPartitionID(strconv.Itoa(int(partition))),
		semconv.MessagingKafkaMessageOffset(int(offset)),
	)
	ps.span.End()
}

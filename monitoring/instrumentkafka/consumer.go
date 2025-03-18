package instrumentkafka

import (
	"context"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/viebiz/lit/monitoring"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// StartConsumeTxn starts a segment for consuming a message from Kafka
func StartConsumeTxn(
	ctx context.Context,
	msg *sarama.ConsumerMessage,
) (context.Context, func(error)) {
	m := monitoring.FromContext(ctx)

	logTags := map[string]string{
		messagingDestinationNameKey:        msg.Topic,
		messagingDestinationPartitionIDKey: strconv.Itoa(int(msg.Partition)),
		messagingKafkaOffsetKey:            strconv.Itoa(int(msg.Offset)),
	}

	// Extract trace context from request headers
	curSpanCtx := otel.GetTextMapPropagator().Extract(ctx, buildConsumeMessageCarrier(msg))
	spanCtx := trace.SpanContextFromContext(curSpanCtx)

	// Start new span
	ctx, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), consumeSpanName,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			semconv.MessagingSystemKafka,
			semconv.MessagingDestinationName(msg.Topic),
			semconv.MessagingKafkaMessageKey(string(msg.Key)),
			semconv.MessagingDestinationPartitionID(strconv.Itoa(int(msg.Partition))),
			semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
		),
	)
	m = monitoring.InjectTracingInfo(m, span.SpanContext())

	m = m.With(logTags)
	ctx = monitoring.SetInContext(ctx, m)

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}
}

// StartCommitSegment starts a segment for committing a message from Kafka
func StartCommitSegment(ctx context.Context, topic, key string, partition int32, offset int64) func() {
	ctx, span := tracer.Start(ctx, commitSpanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.MessagingDestinationName(topic),
			semconv.MessagingKafkaMessageKey(key),
			semconv.MessagingDestinationPartitionID(strconv.Itoa(int(partition))),
			semconv.MessagingKafkaMessageOffset(int(offset)),
		),
	)

	return func() {
		span.End()
	}
}

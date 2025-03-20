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

// StartSyncPublishSegment starts a segment for publishing message to kafka synchronously
func StartSyncPublishSegment(
	ctx context.Context,
	extSvcInfo monitoring.ExternalServiceInfo,
	clientID string,
	pm *sarama.ProducerMessage,
	key string,
) (context.Context, func(partition int32, offset int64, sendErr error)) {
	m := monitoring.FromContext(ctx)

	logTags := map[string]string{
		messagingClientIDKey:        clientID,
		messagingDestinationNameKey: pm.Topic,
		messagingKafkaMessageKey:    key,
	}
	if pm.Partition > 0 {
		logTags[messagingDestinationPartitionIDKey] = strconv.Itoa(int(pm.Partition))
	}

	// Start a new span
	ctx, span := tracer.Start(ctx, produceSpanName,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKafka,
			semconv.MessagingDestinationName(pm.Topic),
			semconv.MessagingDestinationPartitionID(strconv.Itoa(int(pm.Partition))),
			semconv.MessagingClientID(clientID),
			semconv.MessagingKafkaMessageKey(key),
			semconv.MessagingKafkaMessageOffset(int(pm.Offset)),
			semconv.ServerAddress(extSvcInfo.Hostname+":"+extSvcInfo.Port),
		),
	)

	// Inject span context into ProduceMessage headers
	otel.GetTextMapPropagator().Inject(ctx, buildProduceMessageCarrier(pm))

	// Update context with outgoing tracing information and log tags
	ctx = monitoring.SetInContext(ctx,
		monitoring.InjectOutgoingTracingInfo(m, span.SpanContext()).With(logTags),
	)

	ps := buildPublishSegment(span)
	return ctx, ps.End
}

// StartAsyncEnqueueSegment starts a segment for queueing message to async producer
func StartAsyncEnqueueSegment(
	ctx context.Context,
	clientID string,
	pm *sarama.ProducerMessage,
	key string,
) (context.Context, PublishSegment, func()) {
	m := monitoring.FromContext(ctx)

	logTags := map[string]string{
		messagingClientIDKey:        clientID,
		messagingDestinationNameKey: pm.Topic,
		messagingKafkaMessageKey:    key,
	}

	if pm.Partition > 0 {
		logTags[messagingDestinationPartitionIDKey] = strconv.Itoa(int(pm.Partition))
	}

	ctx, span := tracer.Start(ctx, enqueueSpanName, trace.WithSpanKind(trace.SpanKindInternal))
	ctx = monitoring.SetInContext(ctx,
		monitoring.
			InjectOutgoingTracingInfo(m, span.SpanContext()).
			With(logTags),
	)

	// Inject Span Context to request header before send
	otel.GetTextMapPropagator().Inject(ctx, buildProduceMessageCarrier(pm))

	return ctx, buildPublishSegment(span), func() {
		span.End()
	}
}

// StartAsyncPublishSegment starts a segment for publishing message to kafka
func StartAsyncPublishSegment(
	extSvcInfo monitoring.ExternalServiceInfo,
	clientID string,
	pm *sarama.ProducerMessage,
	key string,
) PublishSegment {
	ctx := context.Background()

	// Extract trace context from request headers
	propagator := otel.GetTextMapPropagator()
	curSpanCtx := propagator.Extract(ctx, buildProduceMessageCarrier(pm))
	spanCtx := trace.SpanContextFromContext(curSpanCtx)

	// Start a new span
	ctx, span := tracer.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), produceSpanName,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKafka,
			semconv.MessagingDestinationName(pm.Topic),
			semconv.MessagingDestinationPartitionID(strconv.Itoa(int(pm.Partition))),
			semconv.MessagingClientID(clientID),
			semconv.MessagingKafkaMessageKey(key),
			semconv.MessagingKafkaMessageOffset(int(pm.Offset)),
			semconv.ServerAddress(extSvcInfo.Hostname+":"+extSvcInfo.Port),
		),
	)

	// Inject Span Context to request header before send
	otel.GetTextMapPropagator().Inject(ctx, buildProduceMessageCarrier(pm))

	return buildPublishSegment(span)
}

package instrumentkafka

import (
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName      = "github.com/viebiz/lit/monitoring/instrumentkafka"
	enqueueSpanName = "kafka.enqueue"
	produceSpanName = "kafka.produce"
	consumeSpanName = "kafka.consume"
	commitSpanName  = "kafka.commit"

	// Attributes
	messagingClientIDKey               = "messaging.client.id"
	messagingDestinationNameKey        = "messaging.destination.name"
	messagingDestinationPartitionIDKey = "messaging.destination.partition.id"
	messagingKafkaMessageKey           = "messaging.kafka.message.key"
	messagingKafkaOffsetKey            = "messaging.kafka.offset"
)

var (
	tracer = otel.Tracer(tracerName, trace.WithSchemaURL(semconv.SchemaURL))
)

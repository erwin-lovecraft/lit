package instrumentkafka

import (
	"fmt"
	"strings"

	"github.com/IBM/sarama"
	"github.com/viebiz/lit/monitoring"
)

// InjectAsyncSegmentInfo injects the async segment info into the monitor
func InjectAsyncSegmentInfo(
	m *monitoring.Monitor,
	seg PublishSegment,
	msg *sarama.ProducerMessage,
	sendErr error,
) *monitoring.Monitor {
	logTags := map[string]string{
		messagingDestinationNameKey: msg.Topic,
	}

	if msg.Key != nil {
		if v, err := msg.Key.Encode(); err == nil {
			logTags[messagingKafkaMessageKey] = string(v)
		}
	}

	if sendErr == nil {
		logTags[messagingDestinationPartitionIDKey] = fmt.Sprint(msg.Partition)
		logTags[messagingKafkaOffsetKey] = fmt.Sprint(msg.Offset)
	}

	return monitoring.InjectOutgoingTracingInfo(m, seg.SpanContext()).
		With(logTags)
}

// ExtractSpanIDFromProducerMessage returns the span_id from the trace context header.
// Trace context format is: 00-<trace-id>-<span-id>-<trace-flags>.
// Efficient for AsyncProducer; if the trace context format changes, update this method.
func ExtractSpanIDFromProducerMessage(msg *sarama.ProducerMessage) string {
	const traceparent, sep = "traceparent", "-"
	for _, hdr := range msg.Headers {
		if string(hdr.Key) != traceparent {
			continue
		}
		if parts := strings.Split(string(hdr.Value), sep); len(parts) == 4 {
			return parts[2]
		}
	}
	return ""
}

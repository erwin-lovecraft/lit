package instrumentkafka

import (
	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel/propagation"
)

// produceMessageCarrier is a TextMapCarrier for sarama.ProducerMessage
type produceMessageCarrier struct {
	*sarama.ProducerMessage
}

// Ensure produceMessageCarrier implements TextMapCarrier
var _ propagation.TextMapCarrier = (*produceMessageCarrier)(nil)

func buildProduceMessageCarrier(pm *sarama.ProducerMessage) produceMessageCarrier {
	return produceMessageCarrier{
		ProducerMessage: pm,
	}
}

func (h produceMessageCarrier) Get(key string) string {
	for _, h := range h.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (h produceMessageCarrier) Set(key string, value string) {
	// Remove duplicates by iterating backwards
	for i := 0; i < len(h.Headers); i++ {
		if string(h.Headers[i].Key) == key {
			h.Headers = append(h.Headers[:i], h.Headers[i+1:]...)
			i--
		}
	}
	h.Headers = append(h.Headers, sarama.RecordHeader{
		Key:   []byte(key),
		Value: []byte(value),
	})
}

func (h produceMessageCarrier) Keys() []string {
	out := make([]string, len(h.Headers))
	for i, header := range h.Headers {
		out[i] = string(header.Key)
	}
	return out
}

// consumeMessageCarrier is a TextMapCarrier for sarama.ConsumerMessage
type consumeMessageCarrier struct {
	*sarama.ConsumerMessage
}

// Ensure produceMessageCarrier implements TextMapCarrier
var _ propagation.TextMapCarrier = (*consumeMessageCarrier)(nil)

func buildConsumeMessageCarrier(cm *sarama.ConsumerMessage) consumeMessageCarrier {
	return consumeMessageCarrier{
		ConsumerMessage: cm,
	}
}

func (h consumeMessageCarrier) Get(key string) string {
	for _, h := range h.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (h consumeMessageCarrier) Set(key string, value string) {
	// Remove duplicates by iterating backwards
	for i := 0; i < len(h.Headers); i++ {
		if string(h.Headers[i].Key) == key {
			h.Headers = append(h.Headers[:i], h.Headers[i+1:]...)
			i--
		}
	}
	h.Headers = append(h.Headers, &sarama.RecordHeader{
		Key:   []byte(key),
		Value: []byte(value),
	})
}

func (h consumeMessageCarrier) Keys() []string {
	out := make([]string, len(h.Headers))
	for i, header := range h.Headers {
		out[i] = string(header.Key)
	}
	return out
}

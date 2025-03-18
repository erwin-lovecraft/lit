package kafka

import (
	"crypto/tls"

	"github.com/IBM/sarama"
)

// ProducerOption is a function that sets some option on the producer
type ProducerOption func(*producerConfig)

// ProducerWithAckMode overrides the ack mode of the producer
func ProducerWithAckMode(m AckMode) ProducerOption {
	return func(c *producerConfig) {
		c.Producer.RequiredAcks = sarama.RequiredAcks(m)
	}
}

// ProducerWithAutoCreateTopics sets topic creation to auto
func ProducerWithAutoCreateTopics() ProducerOption {
	return func(c *producerConfig) {
		c.Metadata.AllowAutoTopicCreation = true
	}
}

// ProducerWithTLS sets the TLS config for the producer
func ProducerWithTLS(tlsCfg *tls.Config) ProducerOption {
	return func(c *producerConfig) {
		configWithTLS(c.Config, tlsCfg)
	}
}

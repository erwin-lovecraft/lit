package kafka

import (
	"crypto/tls"
	"time"

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

// ProducerWithRoundRobinPartitioner sets the partitioner to RoundRobin
func ProducerWithRoundRobinPartitioner() ProducerOption {
	return func(c *producerConfig) {
		c.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}
}

// ProducerWithFlushFrequency sets the flush frequency of the producer.
// The best-effort frequency of flushes. Equivalent to `queue.buffering.max.ms` setting of JVM producer
func ProducerWithFlushFrequency(frequency time.Duration) ProducerOption {
	return func(c *producerConfig) {
		c.Producer.Flush.Frequency = frequency
	}
}

// ProducerWithCompression sets the compression codec for the producer
func ProducerWithCompression(codec CompressionCodec) ProducerOption {
	return func(c *producerConfig) {
		c.Producer.Compression = sarama.CompressionCodec(codec)
	}
}

// ProducerWithTLS sets the TLS config for the producer
func ProducerWithTLS(tlsCfg *tls.Config) ProducerOption {
	return func(c *producerConfig) {
		configWithTLS(c.Config, tlsCfg)
	}
}

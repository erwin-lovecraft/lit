package kafka

import (
	"crypto/tls"

	"github.com/IBM/sarama"
)

// ConsumerOption overrides the properties of a consumer
type ConsumerOption func(*consumerConfig)

// ConsumerWithTLS sets the TLS config for the consumer
func ConsumerWithTLS(tlsCfg *tls.Config) ConsumerOption {
	return func(c *consumerConfig) {
		configWithTLS(c.Config, tlsCfg)
	}
}

// ConsumerWithOffsetNewest option with offset newest, default is oldest.
func ConsumerWithOffsetNewest() ConsumerOption {
	return func(c *consumerConfig) {
		c.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
}

// ConsumerMaxRetryPerMessage is the max time we will retry per message
func ConsumerMaxRetryPerMessage(maxRetries int) ConsumerOption {
	return func(c *consumerConfig) {
		c.maxRetriesPerMsg = maxRetries
	}
}

// ConsumerWithAutoCreateTopics sets topic creation to auto
func ConsumerWithAutoCreateTopics() ConsumerOption {
	return func(c *consumerConfig) {
		c.Metadata.AllowAutoTopicCreation = true
	}
}

// ConsumerWithCustomConsumerGroupID sets the group ID to the given ID. This is ONLY to be used for backwards
// compatibility for those clients already using the old consumer group ID in Prod.
func ConsumerWithCustomConsumerGroupID(id string) ConsumerOption {
	return func(c *consumerConfig) {
		c.groupID = id
	}
}

// ConsumerDisablePayloadLogging disables payload logging for the consumer
func ConsumerDisablePayloadLogging() ConsumerOption {
	return func(c *consumerConfig) {
		c.disablePayloadLogging = true
	}
}

package kafka

import (
	"strings"

	"github.com/IBM/sarama"
)

type Config struct {
	AppName string
	Server  string
}

type producerConfig struct {
	*sarama.Config
}

type consumerConfig struct {
	*sarama.Config
	disablePayloadLogging bool
	maxRetriesPerMsg      int
	groupID               string
}

func prepareSaramaConfigBase(cfg Config) *sarama.Config {
	saramaCfg := sarama.NewConfig()
	saramaCfg.RackID = cfg.Server
	saramaCfg.ClientID = cfg.AppName + "." + cfg.Server

	// To meet Sarama's validation requirements, sanitize the client id:-
	saramaCfg.ClientID = strings.Replace(saramaCfg.ClientID, ":", ".", -1)
	saramaCfg.ClientID = strings.Replace(saramaCfg.ClientID, "|", ".", -1)

	saramaCfg.Metadata.AllowAutoTopicCreation = false

	return saramaCfg
}

func prepareProducerConfig(cfg Config) *producerConfig {
	saramaCfg := prepareSaramaConfigBase(cfg)

	// Successfully & failed delivered messages will be returned on the Successes & Errors channel
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true

	return &producerConfig{Config: saramaCfg}
}

func prepareConsumerConfig(cfg Config) *consumerConfig {
	saramaCfg := prepareSaramaConfigBase(cfg)

	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaCfg.Consumer.Return.Errors = true

	return &consumerConfig{
		Config:  saramaCfg,
		groupID: cfg.AppName,

		// This evaluates to around 13hrs with the current backoff config.
		// Ref: consumeBackoff. But we only try up to max of 12hrs
		maxRetriesPerMsg: 35,
	}
}

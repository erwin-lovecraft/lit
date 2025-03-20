package kafka

import (
	"context"
	"errors"

	"github.com/IBM/sarama"
	pkgerrors "github.com/pkg/errors"
	"github.com/viebiz/lit/monitoring"
)

// ConsumerGroup is the kafka consumer
type ConsumerGroup struct {
	monitor  *monitoring.Monitor
	client   sarama.Client
	consumer sarama.ConsumerGroup
	topic    string
	handler  messageHandler
}

type ConsumeHandler func(ctx context.Context, msg ConsumerMessage) error

// NewConsumerGroup creates a new consumer group using the given broker addresses and configuration.
func NewConsumerGroup(
	ctx context.Context,
	cfg Config,
	topic string,
	brokers []string,
	handler ConsumeHandler,
	opts ...ConsumerOption,
) (*ConsumerGroup, error) {
	monitor := monitoring.FromContext(ctx)

	monitor.Infof("Kafka consumer initializing: [%s]", topic)

	// Prepare consumer config
	consumerCfg := prepareConsumerConfig(cfg)
	for _, opt := range opts {
		opt(consumerCfg)
	}

	// Create a new sarama connection
	client, err := sarama.NewClient(brokers, consumerCfg.Config)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "client init failed")
	}

	// Create a new consumer group
	cg, err := sarama.NewConsumerGroupFromClient(consumerCfg.groupID, client)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "creating new consumer")
	}

	newMonitor := monitor.With(map[string]string{
		"kafka.consumer.client_id": consumerCfg.ClientID,
		"kafka.consumer.group_id":  consumerCfg.groupID,
		"kafka.consumer.topic":     topic,
	})

	msgHandler := messageHandler{
		monitor:               newMonitor,
		handler:               handler,
		maxRetriesPerMsg:      consumerCfg.maxRetriesPerMsg,
		disablePayloadLogging: consumerCfg.disablePayloadLogging,
		extSvcInfo:            monitoring.NewExternalServiceInfo(brokers[0]),
	}

	newMonitor.Infof("Kafka consumer initialized. Hostname: [%s], Port: [%s]",
		msgHandler.extSvcInfo.Hostname,
		msgHandler.extSvcInfo.Port,
	)

	return &ConsumerGroup{
		monitor:  newMonitor,
		client:   client,
		consumer: cg,
		topic:    topic,
		handler:  msgHandler,
	}, nil
}

func (c *ConsumerGroup) Consume(ctx context.Context) error {
	consumeErr := make(chan error, 1)
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			if err := c.consumer.Consume(ctx, []string{c.topic}, c.handler); err != nil {
				consumeErr <- pkgerrors.Wrap(err, "consuming failed")
				return
			}
		}
	}()

	select {
	case err := <-consumeErr:
		return err
	case <-ctx.Done():
		c.monitor.Infof("[kafka_consumer] Closing consumer group....")
		if err := c.consumer.Close(); err != nil {
			return pkgerrors.Wrap(err, "stop consumer")
		}
		if err := c.client.Close(); err != nil {
			if !errors.Is(err, sarama.ErrClosedClient) {
				return pkgerrors.Wrap(err, "stop client")
			}
		}
		c.monitor.Infof("[kafka_consumer] ConsumerGroup group closed")
		return nil
	}
}

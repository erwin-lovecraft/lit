package kafka

import (
	"context"
	"errors"

	"github.com/IBM/sarama"
	pkgerrors "github.com/pkg/errors"
	"github.com/viebiz/lit/monitoring"
	"github.com/viebiz/lit/monitoring/instrumentkafka"
)

// SyncProducer publishes Kafka messages, blocking until they have been acknowledged
type SyncProducer struct {
	client     sarama.Client
	producer   sarama.SyncProducer
	extSvcInfo monitoring.ExternalServiceInfo
	clientID   string
}

// NewSyncProducer creates a new sync producer using the given broker addresses and configuration
func NewSyncProducer(
	ctx context.Context,
	cfg Config,
	brokers []string,
	opts ...ProducerOption,
) (SyncProducer, error) {
	monitor := monitoring.FromContext(ctx)

	monitor.Infof("Kafka SyncProducer initializing")

	// Prepare producer config
	producerCfg := prepareProducerConfig(cfg)
	for _, opt := range opts {
		opt(producerCfg)
	}

	// Create a new sarama connection
	client, err := sarama.NewClient(brokers, producerCfg.Config)
	if err != nil {
		return SyncProducer{}, err
	}

	// Create a new sync producer
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return SyncProducer{}, err
	}

	sp := SyncProducer{
		client:     client,
		producer:   p,
		extSvcInfo: monitoring.NewExternalServiceInfo(brokers[0]),
		clientID:   producerCfg.ClientID,
	}

	monitor.Infof("Kafka sync producer initialized. Host [%s], Port: [%s]",
		sp.extSvcInfo.Hostname,
		sp.extSvcInfo.Port,
	)

	return sp, nil
}

func (sp *SyncProducer) SendMessage(ctx context.Context, topic string, payload []byte, opt ProducerMessageOption) error {
	pm, _, err := prepareProducerMessage(topic, payload, opt)
	if err != nil {
		return err
	}

	var (
		partition int32
		offset    int64
	)

	ctx, end := instrumentkafka.StartSyncPublishSegment(ctx, sp.extSvcInfo, sp.clientID, pm, opt.Key)
	defer func() {
		end(partition, offset, err)
	}()
	monitor := monitoring.FromContext(ctx)

	if opt.DisablePayloadLogging {
		monitor.Infof("[kafka_sync_producer] Sending message. Topic: [%s]", topic)
	} else {
		monitor.Infof("[kafka_sync_producer] Sending message. Topic: [%s], Payload: [%s]", topic, payload)
	}

	if _, _, err = sp.producer.SendMessage(pm); err != nil {
		return pkgerrors.WithStack(err)
	}

	monitor.Infof("[kafka_sync_producer] Send message success. Topic: [%s], Partition: [%d], Offset: [%d]", topic, partition, offset)

	return nil
}

func (sp *SyncProducer) Close() error {
	if err := sp.producer.Close(); err != nil {
		return pkgerrors.Wrap(err, "close producer")
	}

	if err := sp.client.Close(); err != nil {
		if !errors.Is(err, sarama.ErrClosedClient) {
			return pkgerrors.Wrap(err, "close sarama client")
		}
	}

	return nil
}

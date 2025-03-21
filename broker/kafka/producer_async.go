package kafka

import (
	"context"
	"errors"

	"github.com/IBM/sarama"
	pkgerrors "github.com/pkg/errors"
	"github.com/viebiz/lit/monitoring"
	"github.com/viebiz/lit/monitoring/instrumentkafka"
)

// AsyncProducer publishes Kafka messages using a non-blocking API
type AsyncProducer struct {
	monitor      *monitoring.Monitor
	client       sarama.Client
	producer     sarama.AsyncProducer
	extSvcInfo   monitoring.ExternalServiceInfo
	clientID     string
	publishQueue chan asyncMessage                         // This is needed to make ops on instSegments thread safe
	instSegments map[string]instrumentkafka.PublishSegment // Cache of instrumentation segments, map message spanID to segment
}

// NewAsyncProducer creates a new async producer using the given broker addresses and configuration
func NewAsyncProducer(
	ctx context.Context,
	cfg Config,
	brokers []string,
	opts ...ProducerOption,
) (*AsyncProducer, error) {
	monitor := monitoring.FromContext(ctx)

	monitor.Infof("Kafka AsyncProducer initializing")

	// Prepare producer config
	producerCfg := prepareProducerConfig(cfg)
	for _, opt := range opts {
		opt(producerCfg)
	}

	// Create a new sarama connection
	client, err := sarama.NewClient(brokers, producerCfg.Config)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "init client")
	}

	// Create a new async producer
	p, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "init async producer")
	}

	ap := &AsyncProducer{
		monitor:      monitor,
		client:       client,
		producer:     p,
		extSvcInfo:   monitoring.NewExternalServiceInfo(brokers[0]),
		clientID:     producerCfg.ClientID,
		publishQueue: make(chan asyncMessage), // Must be unbuffered
		instSegments: make(map[string]instrumentkafka.PublishSegment),
	}

	monitor.Infof("Kafka AsyncProducer initialized. Host: [%s], Port: [%s]",
		ap.extSvcInfo.Hostname,
		ap.extSvcInfo.Port,
	)

	return ap, nil
}

func (ap *AsyncProducer) SendMessage(ctx context.Context, topic string, payload []byte, opt ProducerMessageOption) error {
	pm, messageKey, err := prepareProducerMessage(topic, payload, opt)
	if err != nil {
		return err
	}

	ctx, seg, end := instrumentkafka.StartAsyncEnqueueSegment(ctx, ap.clientID, pm, opt.Key)
	defer end()
	monitor := monitoring.FromContext(ctx)

	if opt.DisablePayloadLogging {
		monitor.Infof("[kafka_async_producer] Euqueue message. Topic: [%s]", topic)
	} else {
		monitor.Infof("[kafka_async_producer] Enqueue message. Topic: [%s], Payload: [%s]", topic, payload)
	}

	ap.publishQueue <- asyncMessage{
		key:         messageKey,
		producerMsg: pm,
		segment:     seg,
	}

	return nil
}

func (ap *AsyncProducer) handleAsyncMsg(msg *sarama.ProducerMessage, sendErr error) {
	if msg == nil {
		return // Should never happen but just to be safe
	}

	spanID := instrumentkafka.ExtractSpanIDFromProducerMessage(msg)
	if spanID == "" {
		return // Skip if no spanID found
	}

	var seg instrumentkafka.PublishSegment
	if s, exists := ap.instSegments[spanID]; exists {
		delete(ap.instSegments, spanID)
		seg = s
		seg.End(msg.Partition, msg.Offset, sendErr)
	}

	monitor := instrumentkafka.InjectAsyncSegmentInfo(ap.monitor, seg, msg, sendErr)
	if sendErr != nil {
		monitor.Errorf(pkgerrors.WithStack(sendErr), "[kafka_async_producer] Send Error")
	} else {
		monitor.Infof("[kafka_async_producer] Send Success. Partition: [%d], Offset: [%d]", msg.Partition, msg.Offset)
	}
}

func (ap *AsyncProducer) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			ap.monitor.Infof("[kafka_async_producer] Closing listener....")
			if err := ap.producer.Close(); err != nil {
				return pkgerrors.Wrap(err, "close producer")
			}
			if err := ap.client.Close(); err != nil {
				if !errors.Is(err, sarama.ErrClosedClient) {
					return pkgerrors.Wrap(err, "close sarama client")
				}
			}

			ap.monitor.Infof("[kafka_async_producer] Listener closed")
			return nil

		case msg := <-ap.publishQueue:
			ps := instrumentkafka.StartAsyncPublishSegment(
				ap.extSvcInfo,
				ap.clientID,
				msg.producerMsg,
				msg.key,
			)
			ap.instSegments[ps.SpanID()] = ps
			ap.monitor.Infof("[kafka_async_producer] Publish message")
			ap.producer.Input() <- msg.producerMsg

		case msg := <-ap.producer.Successes():
			ap.handleAsyncMsg(msg, nil)
		case err := <-ap.producer.Errors():
			ap.handleAsyncMsg(err.Msg, err.Err)
		}
	}
}

type asyncMessage struct {
	key         string // Keep the plain key for easy access
	producerMsg *sarama.ProducerMessage
	segment     instrumentkafka.PublishSegment
}

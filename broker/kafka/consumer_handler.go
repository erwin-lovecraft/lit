package kafka

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/IBM/sarama"
	"github.com/cenkalti/backoff/v4"
	"github.com/viebiz/lit/monitoring"
	"github.com/viebiz/lit/monitoring/instrumentkafka"
)

// ConsumerMessage encapsulates a Kafka message returned by the consumer.
type ConsumerMessage struct {
	ID      ConsumerMessageID
	Value   []byte
	Headers map[string]string
}

// ConsumerMessageID is the unique identifier of the message
type ConsumerMessageID struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       string
}

type messageHandler struct {
	monitor               *monitoring.Monitor
	handler               ConsumeHandler
	disablePayloadLogging bool
	maxRetriesPerMsg      int
	extSvcInfo            monitoring.ExternalServiceInfo
}

func (h messageHandler) Setup(s sarama.ConsumerGroupSession) error {
	h.monitor.
		WithTag("kafka_consumer_member_id", s.MemberID()).
		Infof("[Kafka ConsumerGroup] ConsumerGroup Ready. GenerationID: [%d], Member ID: [%s], Partition Allocation: [%v], ", s.GenerationID(), s.MemberID(), s.Claims())
	return nil
}

func (h messageHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.monitor.Infof("[Kafka ConsumerGroup] Cleaning up...")
	return nil
}

func (h messageHandler) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	for msg := range c.Messages() {
		h.consume(s, msg)
	}
	return nil
}

func (h messageHandler) consume(s sarama.ConsumerGroupSession, cm *sarama.ConsumerMessage) {
	ctx, segEnd := instrumentkafka.StartConsumeTxn(monitoring.SetInContext(context.Background(), h.monitor), cm)
	monitor := monitoring.FromContext(ctx)

	var msgKey string
	if cm.Key != nil {
		msgKey = string(cm.Key)
	}

	var err error
	defer func() {
		if rcv := recover(); rcv != nil {
			err = fmt.Errorf("panic err: %s", rcv)
			monitoring.FromContext(ctx).Errorf(err, "Caught PANIC. Stack Trace: %s", debug.Stack())
		}

		segEnd(err)
	}()

	if h.disablePayloadLogging {
		monitor.Infof("[Kafka ConsumerGroup] Consuming: Partition: [%d], Offset: [%d]", cm.Partition, cm.Offset)
	} else {
		monitor.Infof("[Kafka ConsumerGroup] Consuming: Partition: [%d], Offset: [%d], Payload: [%s]", cm.Partition, cm.Offset, cm.Value)
	}

	msg := ConsumerMessage{
		ID: ConsumerMessageID{
			Topic:     cm.Topic,
			Partition: cm.Partition,
			Offset:    cm.Offset,
			Key:       msgKey,
		},
		Value:   cm.Value,
		Headers: make(map[string]string, len(cm.Headers)), // It's ok to possibly over provision in case of duplicate.
	}
	if cm.Headers != nil {
		for _, r := range cm.Headers {
			msg.Headers[string(r.Key[:])] = string(r.Value[:])
		}
	}

	var attempts int

	if err = backoff.Retry(func() error {
		attempts++
		monitor.Infof("[Kafka ConsumerGroup] Consuming Attempt: [%d]", attempts)

		if err = h.handler(ctx, msg); err != nil {
			monitor.Errorf(err, "consume message failed")
			return err
		}
		return nil
	}, backoff.WithContext(consumeBackoff(h.maxRetriesPerMsg), ctx)); err != nil {
		monitor.Infof("[Kafka ConsumerGroup] Giving up on processing. Partition: [%d], Offset: [%d] after [%d] attempts. Will just commit and move on", cm.Partition, cm.Offset, attempts)
	}

	h.commitMessageOffset(ctx, s, msg.ID)

	monitor.Infof("[Kafka ConsumerGroup] Consumed: Partition: [%d], Offset: [%d]", cm.Partition, cm.Offset)
}

// CommitMessageOffset commits the message's offset+1 for the topic & partition
func (h messageHandler) commitMessageOffset(
	ctx context.Context,
	cgs sarama.ConsumerGroupSession,
	msgID ConsumerMessageID,
) {
	offsetToCommit := msgID.Offset + 1 // Should always commit next offset as best practice

	end := instrumentkafka.StartCommitSegment(ctx, msgID.Topic, msgID.Key, msgID.Partition, msgID.Offset)
	defer end()

	monitoring.FromContext(ctx).Infof("[kafka_consumer] Committing offset: [%d]", offsetToCommit)

	cgs.MarkOffset(msgID.Topic, msgID.Partition, offsetToCommit, "")

	monitoring.FromContext(ctx).Infof("[kafka_consumer] Committing offset complete for: [%d]", offsetToCommit)
}

func consumeBackoff(maxRetries int) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 5 * time.Second
	b.RandomizationFactor = 0
	b.Multiplier = 1.25
	b.MaxInterval = 30 * time.Minute
	b.MaxElapsedTime = 12 * time.Hour
	return backoff.WithMaxRetries(b, uint64(maxRetries))
}

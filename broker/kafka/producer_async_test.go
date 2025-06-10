package kafka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAsyncProducer(t *testing.T) {
	// Setup msg prop
	nowUnix := time.Now().Unix()
	givenMessageKey := fmt.Sprintf("key-%d", nowUnix)
	givenTopic := fmt.Sprintf("topic-%d", nowUnix)
	givenMessage := fmt.Sprintf("message-%d", nowUnix)

	kafkaURL := getKafkaURL()

	// Setup producer
	producer, err := NewAsyncProducer(
		context.Background(),
		Config{AppName: "lit", Server: "local"},
		[]string{kafkaURL},
		ProducerWithAutoCreateTopics(),
	)
	require.NoError(t, err)

	// Setup msg handler
	msgHandler := func(ctx context.Context, msg ConsumerMessage) error {
		require.Equal(t, msg.ID.Key, givenMessageKey)
		require.Equal(t, msg.ID.Topic, givenTopic)
		require.Equal(t, string(msg.Value), givenMessage)
		return nil
	}
	// Setup consumer
	consumer, err := NewConsumerGroup(
		context.Background(),
		Config{AppName: "lit", Server: "local"},
		[]string{givenTopic},
		[]string{kafkaURL},
		msgHandler,
		ConsumerWithAutoCreateTopics(),
	)
	require.NoError(t, err)

	ctx := context.Background()
	// When
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func(childCtx context.Context) {
		require.NoError(t, producer.Listen(childCtx))
	}(timeoutCtx)

	// When: producer sends msg to topic
	err = producer.SendMessage(
		ctx,
		givenTopic,
		[]byte(givenMessage),
		ProducerMessageOption{Key: givenMessageKey},
	)
	require.NoError(t, err)

	// Then:
	require.NoError(t, consumer.Consume(timeoutCtx))
}

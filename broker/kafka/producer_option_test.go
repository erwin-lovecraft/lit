package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewProducerWithOptions(t *testing.T) {
	ctx := context.Background()

	p, err := NewSyncProducer(
		ctx, Config{
			AppName: "lightning",
			Server:  "test",
		},
		[]string{getKafkaURL()},
		ProducerWithAutoCreateTopics(),
		ProducerWithRoundRobinPartitioner(),
		ProducerWithFlushFrequency(500*time.Millisecond), // flush every 500ms
		ProducerWithCompression(CompressionSnappy),
	)
	require.NoError(t, err)
	require.NotNil(t, p)
}

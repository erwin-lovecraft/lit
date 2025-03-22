package instrumentkafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
)

func TestProducerMessageCarrier_Get(t *testing.T) {
	// GIVEN
	msg := &sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("traceparent"), Value: []byte("test-value")},
			{Key: []byte("another-key"), Value: []byte("another-value")},
		},
	}
	carrier := buildProduceMessageCarrier(msg)

	// WHEN
	value := carrier.Get("traceparent")
	emptyValue := carrier.Get("non-existent")

	// THEN
	require.Equal(t, "test-value", value)
	require.Equal(t, "", emptyValue)
}

func TestProducerMessageCarrier_Set(t *testing.T) {
	// GIVEN
	msg := &sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("existing-key"), Value: []byte("old-value")},
		},
	}
	carrier := buildProduceMessageCarrier(msg)

	// WHEN - Adding new key
	carrier.Set("traceparent", "new-value")

	// THEN - Should have both headers
	require.Len(t, carrier.Headers, 2)
	require.Equal(t, "old-value", carrier.Get("existing-key"))
	require.Equal(t, "new-value", carrier.Get("traceparent"))

	// WHEN - Updating existing key
	carrier.Set("existing-key", "updated-value")

	// THEN - Should have same length but updated value
	require.Len(t, carrier.Headers, 2)
	require.Equal(t, "updated-value", carrier.Get("existing-key"))
	require.Equal(t, "new-value", carrier.Get("traceparent"))
}

func TestProducerMessageCarrier_Set_RemoveDuplicates(t *testing.T) {
	// GIVEN
	msg := &sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("traceparent"), Value: []byte("value1")},
			{Key: []byte("other-key"), Value: []byte("other-value")},
			{Key: []byte("traceparent"), Value: []byte("value2")},
		},
	}
	carrier := buildProduceMessageCarrier(msg)

	// WHEN
	carrier.Set("traceparent", "new-value")

	// THEN
	require.Len(t, carrier.Headers, 2)
	require.Equal(t, "new-value", carrier.Get("traceparent"))
	require.Equal(t, "other-value", carrier.Get("other-key"))
}

func TestProducerMessageCarrier_Keys(t *testing.T) {
	// GIVEN
	msg := &sarama.ProducerMessage{
		Headers: []sarama.RecordHeader{
			{Key: []byte("key1"), Value: []byte("value1")},
			{Key: []byte("key2"), Value: []byte("value2")},
			{Key: []byte("key3"), Value: []byte("value3")},
		},
	}
	carrier := buildProduceMessageCarrier(msg)

	// WHEN
	keys := carrier.Keys()

	// THEN
	require.ElementsMatch(t, []string{"key1", "key2", "key3"}, keys)
}

func TestProducerMessageCarrier_EmptyHeaders(t *testing.T) {
	// GIVEN
	msg := &sarama.ProducerMessage{}
	carrier := buildProduceMessageCarrier(msg)

	// WHEN/THEN
	require.Equal(t, "", carrier.Get("anything"))
	require.Empty(t, carrier.Keys())

	// WHEN
	carrier.Set("new-key", "new-value")

	// THEN
	require.Equal(t, "new-value", carrier.Get("new-key"))
	require.ElementsMatch(t, []string{"new-key"}, carrier.Keys())
}

func TestConsumeMessageCarrier_Get(t *testing.T) {
	// GIVEN
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("traceparent"), Value: []byte("test-value")},
			{Key: []byte("another-key"), Value: []byte("another-value")},
		},
	}
	carrier := buildConsumeMessageCarrier(msg)

	// WHEN
	value := carrier.Get("traceparent")
	emptyValue := carrier.Get("non-existent")

	// THEN
	require.Equal(t, "test-value", value)
	require.Equal(t, "", emptyValue)
}

func TestConsumeMessageCarrier_Set(t *testing.T) {
	// GIVEN
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("existing-key"), Value: []byte("old-value")},
		},
	}
	carrier := buildConsumeMessageCarrier(msg)

	// WHEN - Adding new key
	carrier.Set("traceparent", "new-value")

	// THEN - Should have both headers
	require.Len(t, carrier.Headers, 2)
	require.Equal(t, "old-value", carrier.Get("existing-key"))
	require.Equal(t, "new-value", carrier.Get("traceparent"))

	// WHEN - Updating existing key
	carrier.Set("existing-key", "updated-value")

	// THEN - Should have same length but updated value
	require.Len(t, carrier.Headers, 2)
	require.Equal(t, "updated-value", carrier.Get("existing-key"))
	require.Equal(t, "new-value", carrier.Get("traceparent"))
}

func TestConsumeMessageCarrier_Set_RemoveDuplicates(t *testing.T) {
	// GIVEN
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("traceparent"), Value: []byte("value1")},
			{Key: []byte("other-key"), Value: []byte("other-value")},
			{Key: []byte("traceparent"), Value: []byte("value2")},
		},
	}
	carrier := buildConsumeMessageCarrier(msg)

	// WHEN
	carrier.Set("traceparent", "new-value")

	// THEN
	require.Len(t, carrier.Headers, 2)
	require.Equal(t, "new-value", carrier.Get("traceparent"))
	require.Equal(t, "other-value", carrier.Get("other-key"))
}

func TestConsumeMessageCarrier_Keys(t *testing.T) {
	// GIVEN
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("key1"), Value: []byte("value1")},
			{Key: []byte("key2"), Value: []byte("value2")},
			{Key: []byte("key3"), Value: []byte("value3")},
		},
	}
	carrier := buildConsumeMessageCarrier(msg)

	// WHEN
	keys := carrier.Keys()

	// THEN
	require.ElementsMatch(t, []string{"key1", "key2", "key3"}, keys)
}

func TestConsumeMessageCarrier_EmptyHeaders(t *testing.T) {
	// GIVEN
	msg := &sarama.ConsumerMessage{}
	carrier := buildConsumeMessageCarrier(msg)

	// WHEN/THEN
	require.Equal(t, "", carrier.Get("anything"))
	require.Empty(t, carrier.Keys())

	// WHEN
	carrier.Set("new-key", "new-value")

	// THEN
	require.Equal(t, "new-value", carrier.Get("new-key"))
	require.ElementsMatch(t, []string{"new-key"}, carrier.Keys())
}

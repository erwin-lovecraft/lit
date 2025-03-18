package kafka

import (
	"github.com/IBM/sarama"
)

type ProducerMessageOption struct {
	Key                   string
	Partition             *int32 // Cannot use int because of forced conversion
	Headers               map[string]string
	DisablePayloadLogging bool
}

func prepareProducerMessage(topic string, payload []byte, opt ProducerMessageOption) (*sarama.ProducerMessage, string, error) {
	if topic == "" {
		return nil, "", ErrEmptyTopic
	}
	if len(payload) == 0 {
		return nil, "", ErrNoPayloadProvided
	}

	// Generate a new UUID if key is empty
	if opt.Key == "" {
		opt.Key = generateIDFunc()
	}

	pm := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(payload), Key: sarama.StringEncoder(opt.Key)}

	// Set partition if provided
	if opt.Partition != nil {
		pm.Partition = *opt.Partition
	}

	// Set headers if provided
	if l := len(opt.Headers); l > 0 {
		pm.Headers = make([]sarama.RecordHeader, 0, l)
		for k, v := range opt.Headers {
			pm.Headers = append(pm.Headers, sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
	}

	return pm, opt.Key, nil
}

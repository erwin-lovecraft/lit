package kafka

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
)

func TestPrepareProducerMessage(t *testing.T) {
	i := int32(3)
	type arg struct {
		givenTopic   string
		givenPayload []byte
		givenOpt     ProducerMessageOption
		expErr       error
		expMsg       *sarama.ProducerMessage
	}
	tcs := map[string]arg{
		"all options": {
			givenTopic:   "t1",
			givenPayload: []byte("pay"),
			givenOpt: ProducerMessageOption{
				Key:       "key1",
				Partition: &i,
				Headers:   map[string]string{"k1": "v1", "k2": "v2"},
			},
			expMsg: &sarama.ProducerMessage{
				Topic:     "t1",
				Key:       sarama.StringEncoder("key1"),
				Value:     sarama.ByteEncoder("pay"),
				Headers:   []sarama.RecordHeader{{Key: []byte("k1"), Value: []byte("v1")}, {Key: []byte("k2"), Value: []byte("v2")}},
				Partition: i,
			},
		},
		"no options": {
			givenTopic:   "t1",
			givenPayload: []byte("pay"),
			expMsg: &sarama.ProducerMessage{
				Topic: "t1",
				Key:   sarama.StringEncoder("uid"),
				Value: sarama.ByteEncoder("pay"),
			},
		},
		"no topic": {
			givenPayload: []byte("pay"),
			expErr:       ErrEmptyTopic,
		},
		"no payload": {
			givenTopic: "t1",
			expErr:     errors.New("no payload provided"),
		},
	}
	for desc, tc := range tcs {
		t.Run(desc, func(t *testing.T) {
			// Given:
			staticUID := "uid"
			generateIDFunc = func() string { return staticUID }
			defer func() {
				generateIDFunc = generateID
			}()

			// When:
			pm, key, err := prepareProducerMessage(tc.givenTopic, tc.givenPayload, tc.givenOpt)

			// Then:
			require.Equal(t, tc.expErr, err)
			if tc.expErr != nil {
				require.Nil(t, pm)
				return
			}
			require.Equal(t, tc.expMsg.Topic, pm.Topic)
			require.Equal(t, tc.expMsg.Partition, pm.Partition)
			require.Equal(t, tc.expMsg.Headers, pm.Headers)
			require.Equal(t, tc.expMsg.Offset, pm.Offset)
			require.Equal(t, tc.expMsg.Timestamp, pm.Timestamp)

			expKeyBytes, err := tc.expMsg.Key.Encode()
			require.NoError(t, err)
			pmKeyBytes, err := pm.Key.Encode()
			require.NoError(t, err)
			require.Equal(t, expKeyBytes, pmKeyBytes)
			require.Equal(t, expKeyBytes, []byte(key))

			expValBytes, err := tc.expMsg.Value.Encode()
			require.NoError(t, err)
			pmValBytes, err := pm.Value.Encode()
			require.NoError(t, err)
			require.Equal(t, expValBytes, pmValBytes)
		})
	}
}

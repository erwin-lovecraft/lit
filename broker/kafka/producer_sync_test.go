package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSyncProducer(t *testing.T) {
	type arg struct {
		givenBrokerURL       string
		givenAppCfg          Config
		givenServerName      string
		givenProducerOptions []ProducerOption
		expErr               error
	}
	tcs := map[string]arg{
		"invalid url": {
			givenBrokerURL:  "notvalid",
			givenAppCfg:     Config{AppName: "lit", Server: "local"},
			givenServerName: "svr",
			expErr:          errors.New("kafka: client has run out of available brokers to talk to: dial tcp: address notvalid: missing port in address"),
		},
		"broker unreachable": {
			givenBrokerURL:  "notvalid:9092",
			givenAppCfg:     Config{AppName: "lit", Server: "local"},
			givenServerName: "svr",
			expErr:          errors.New("kafka: client has run out of available brokers to talk to: dial tcp: lookup notvalid"),
		},
		"success": {
			givenBrokerURL:  getKafkaURL(),
			givenAppCfg:     Config{AppName: "lit", Server: "local"},
			givenServerName: "svr",
		},
		"success with ack mode none": {
			givenBrokerURL:  getKafkaURL(),
			givenAppCfg:     Config{AppName: "lit", Server: "local"},
			givenServerName: "svr",
			givenProducerOptions: []ProducerOption{
				ProducerWithAckMode(AckModeNone),
			},
		},
		"success with auto create topics": {
			givenBrokerURL:  getKafkaURL(),
			givenAppCfg:     Config{AppName: "lit", Server: "local"},
			givenServerName: "svr",
			givenProducerOptions: []ProducerOption{
				ProducerWithAutoCreateTopics(),
			},
		},
	}

	for desc, tc := range tcs {
		tc := tc
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			// Given
			// When
			sp, err := NewSyncProducer(
				context.Background(),
				tc.givenAppCfg,
				[]string{tc.givenBrokerURL},
				tc.givenProducerOptions...)

			// Then
			if tc.expErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErr.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, sp)
				require.NotNil(t, sp.client)
				require.NotNil(t, sp.producer)
			}
		})
	}
}

func TestProducerClose(t *testing.T) {
	sp, err := NewSyncProducer(
		context.Background(),
		Config{AppName: "lit", Server: "local"},
		[]string{getKafkaURL()},
	)
	require.NoError(t, err)
	require.NoError(t, sp.Close())
}

package kafka

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func getKafkaURL() string {
	if url := os.Getenv("KAFKA_URL"); url != "" {
		return url
	}
	return "localhost:9092"
}

func TestNewConsumer(t *testing.T) {
	type arg struct {
		givenBrokerURL            string
		givenAppCfg               Config
		givenTopic                string
		givenConsumerOptions      []ConsumerOption
		expErr                    error
		expPayloadLoggingDisabled bool
	}
	tcs := map[string]arg{
		"invalid url": {
			givenBrokerURL: "notvalid",
			givenAppCfg:    Config{AppName: "lit", Server: "local"},
			givenTopic:     "athena",
			expErr:         errors.New("client init failed: kafka: client has run out of available brokers to talk to: dial tcp: address notvalid: missing port in address"),
		},
		"broker unreachable": {
			givenBrokerURL: "notvalid:9092",
			givenAppCfg:    Config{AppName: "lit", Server: "local"},
			givenTopic:     "athena",
			expErr:         errors.New("client init failed: kafka: client has run out of available brokers to talk to"),
		},
		"success": {
			givenBrokerURL: getKafkaURL(),
			givenAppCfg:    Config{AppName: "lit", Server: "local"},
			givenTopic:     "athena",
		},
		"success with consumer option": {
			givenBrokerURL: getKafkaURL(),
			givenAppCfg:    Config{AppName: "lit", Server: "local"},
			givenTopic:     "athena",
			givenConsumerOptions: []ConsumerOption{
				ConsumerWithOffsetNewest(),
			},
		},
		"success with custom group ID": {
			givenBrokerURL: getKafkaURL(),
			givenAppCfg:    Config{AppName: "lit", Server: "local"},
			givenTopic:     "athena",
			givenConsumerOptions: []ConsumerOption{
				ConsumerWithCustomConsumerGroupID("project-app_subc"),
			},
		},
		"success with disabled payload logging": {
			givenBrokerURL: getKafkaURL(),
			givenAppCfg:    Config{AppName: "lit", Server: "local"},
			givenTopic:     "athena",
			givenConsumerOptions: []ConsumerOption{
				ConsumerDisablePayloadLogging(),
			},
			expPayloadLoggingDisabled: true,
		},
	}

	for desc, tc := range tcs {
		tc := tc
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			// Given
			// When
			ctx := context.Background()

			c, err := NewConsumerGroup(
				ctx,
				tc.givenAppCfg,
				tc.givenTopic,
				[]string{tc.givenBrokerURL},
				func(ctx context.Context, msg ConsumerMessage) error {
					return nil
				},
				tc.givenConsumerOptions...)

			if tc.expErr != nil {
				require.Error(t, err)
				require.True(t, strings.HasPrefix(err.Error(), tc.expErr.Error()))
			} else {
				require.NoError(t, err)
				require.NotNil(t, c)
				require.NotNil(t, c.consumer)
				require.NotNil(t, c.client)
				require.Equalf(t, tc.expPayloadLoggingDisabled, c.handler.disablePayloadLogging, "payload logging not set to correct value")
				require.Equal(t, c.topic, tc.givenTopic)
			}
		})
	}
}

func TestConsume(t *testing.T) {
	type arg struct {
		givenTopic  string
		givenOption ConsumerOption
		mockErr     error
	}
	tcs := map[string]arg{
		"handle msg success": {
			givenTopic:  fmt.Sprintf("success-%d", time.Now().Unix()),
			givenOption: ConsumerMaxRetryPerMessage(1),
		},
		"handle msg err - should not return err": {
			givenTopic:  fmt.Sprintf("failure-%d", time.Now().Unix()),
			givenOption: ConsumerMaxRetryPerMessage(1),
			mockErr:     errors.New("something wrong"),
		},
	}
	for desc, tc := range tcs {
		t.Run(desc, func(t *testing.T) {
			kafkaURL := getKafkaURL()

			// Setup msg prop
			nowUnix := time.Now().Unix()
			givenMessageKey := fmt.Sprintf("key-%d", nowUnix)
			givenHeaders := map[string]string{"test-key": "test-val"}

			// Setup producer
			producer, err := NewSyncProducer(
				context.Background(),
				Config{AppName: "lit", Server: "local"},
				[]string{kafkaURL},
				ProducerWithAutoCreateTopics(),
			)
			require.NoError(t, err)

			// Setup msg handler
			msgHandler := func(ctx context.Context, msg ConsumerMessage) error {
				require.Equal(t, msg.ID.Key, givenMessageKey)
				for k, v := range givenHeaders {
					require.Equal(t, v, msg.Headers[k])
				}
				return tc.mockErr
			}
			// Setup consumer
			consumer, err := NewConsumerGroup(
				context.Background(),
				Config{AppName: "lit", Server: "local"},
				tc.givenTopic,
				[]string{kafkaURL},
				msgHandler,
				tc.givenOption,
			)
			require.NoError(t, err)

			ctx := context.Background()
			// Given
			// When
			err = producer.SendMessage(ctx, tc.givenTopic, []byte("hello"), ProducerMessageOption{Key: givenMessageKey, Headers: givenHeaders})
			require.NoError(t, err)
			require.NoError(t, producer.Close())

			// When: consume msg
			timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			require.NoError(t, consumer.Consume(timeoutCtx))
		})
	}
}

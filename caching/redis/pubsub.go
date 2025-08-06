package redis

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/viebiz/lit/monitoring"
)

// Publish posts a new message to server
func (client redisClient) Publish(ctx context.Context, channel string, payload any) error {
	// Use SPUBLISH for better efficiency with sharding (Redis Cluster), If not using sharding it the same PUBLISH
	return client.rdb.SPublish(ctx, channel, payload).Err()
}

func (client redisClient) Subscribe(ctx context.Context, channels []string, handler MessageHandler) Subscriber {
	monitor := monitoring.FromContext(ctx)
	monitor.Infof("Redis subscriber initializing: [%s]", channels)

	monitor = monitor.With(map[string]string{
		"redis.subscriber.channels": strings.Join(channels, ","),
	})

	return &subscriber{
		rdb:      client.rdb,
		channels: channels,
		handler:  handler,
		monitor:  monitoring.FromContext(ctx),
	}
}

type Message struct {
	Channel      string
	Pattern      string
	Payload      string
	PayloadSlice []string
}

func (msg *Message) from(m *redis.Message) {
	msg.Channel = m.Channel
	msg.Pattern = m.Pattern
	msg.Payload = m.Payload
	msg.PayloadSlice = m.PayloadSlice
}

type MessageHandler func(ctx context.Context, msg Message) error

type Subscriber interface {
	Subscribe(ctx context.Context) error
}

type subscriber struct {
	rdb      redis.UniversalClient
	monitor  *monitoring.Monitor
	channels []string
	handler  MessageHandler
}

func (s *subscriber) Subscribe(ctx context.Context) error {
	subErr := make(chan error)
	pubsub := s.rdb.SSubscribe(ctx, s.channels...)
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				subErr <- pkgerrors.Wrap(err, "receive message failed")
				return
			}

			s.handleMessage(msg)
		}
	}()

	select {
	case err := <-subErr:
		return err
	case <-ctx.Done():
		s.monitor.Infof("[redis_subscriber] Closing subscriber channels: %v", s.channels)
		if err := pubsub.Close(); err != nil {
			return pkgerrors.Wrap(err, "closing subscriber failed")
		}
		s.monitor.Infof("[redis_subscriber] Subscriber closed")
		return nil
	}
}

func (s *subscriber) handleMessage(msg *redis.Message) {
	ctx := context.Background()

	var err error
	defer func() {
		if rcv := recover(); rcv != nil {
			err = fmt.Errorf("panic: %v", rcv)
			monitoring.FromContext(ctx).Errorf(err, "Caught PANIC. Stack trace: %s", debug.Stack())
		}
	}()

	s.monitor.Infof("[redis_subscriber] Received message: %s", msg.Payload)
	var m Message
	m.from(msg)

	// TODO: Implement retry backoff
	if err := s.handler(ctx, m); err != nil {
		s.monitor.Errorf(err, "[redis_subscriber] Handle message failed")
		return
	}
}

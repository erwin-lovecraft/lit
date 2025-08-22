package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type ChannelOption struct {
	ChannelSize         int
	HealthCheckInterval time.Duration
	SendTimeout         time.Duration
}

func (opt ChannelOption) toRedisOptions() []redis.ChannelOption {
	opts := make([]redis.ChannelOption, 0, 3)
	if opt.ChannelSize > 0 {
		opts = append(opts, redis.WithChannelSize(opt.ChannelSize))
	}
	if opt.HealthCheckInterval > 0 {
		opts = append(opts, redis.WithChannelHealthCheckInterval(opt.HealthCheckInterval))
	}
	if opt.SendTimeout > 0 {
		opts = append(opts, redis.WithChannelSendTimeout(opt.SendTimeout))
	}
	return opts
}

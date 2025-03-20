package kafka

import (
	"crypto/tls"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

var (
	generateIDFunc = generateID
)

func configWithTLS(c *sarama.Config, t *tls.Config) {
	c.Net.TLS.Enable = true
	c.Net.TLS.Config = t
}

func generateID() string {
	uid, err := uuid.NewRandom()
	if err != nil {
		return strconv.Itoa(int(time.Now().Unix()))
	}
	return uid.String()
}

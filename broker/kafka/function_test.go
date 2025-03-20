package kafka

import (
	"crypto/tls"
	"strconv"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestConfigWithTLS(t *testing.T) {
	cfg := sarama.NewConfig()
	require.False(t, cfg.Net.TLS.Enable)

	dummyTLS := &tls.Config{
		ServerName: "example.com",
	}

	configWithTLS(cfg, dummyTLS)
	require.True(t, cfg.Net.TLS.Enable)
	require.Equal(t, dummyTLS, cfg.Net.TLS.Config)
}

func TestGenerateID(t *testing.T) {
	id := generateID()
	require.NotEmpty(t, id)

	_, err := uuid.Parse(id)
	if err != nil {
		_, convErr := strconv.Atoi(id)
		require.NoError(t, convErr)
	}

	if _, convErr := strconv.Atoi(id); convErr == nil {
		currentUnix := int(time.Now().Unix())
		idInt, _ := strconv.Atoi(id)
		// Accept a difference of a few seconds due to execution time.
		require.InDelta(t, currentUnix, idInt, 5)
	}
}

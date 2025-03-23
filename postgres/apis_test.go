package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/monitoring"
)

func getPGURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgresql://lit:lit@localhost:54321/master?sslmode=disable"
}

func TestNewPool(t *testing.T) {
	ctx := context.Background()
	m, err := monitoring.New(monitoring.Config{})
	require.NoError(t, err)

	ctx = monitoring.SetInContext(ctx, m)
	dbURL := getPGURL()

	pool, err := NewPool(ctx, dbURL,
		1, 1,
		PoolMaxConnLifetime(1),
		AttemptPingUponStartup(),
	)
	require.NoError(t, err)
	require.NotNil(t, pool)
}

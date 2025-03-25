package instrumentpg

import (
	"github.com/viebiz/lit/postgres"
)

// WithInstrumentation wraps a BeginnerExecutor with instrumentation.
func WithInstrumentation(pool postgres.BeginnerExecutor) postgres.BeginnerExecutor {
	return instrumentedDB{BeginnerExecutor: pool}
}

// WithInstrumentationTx wraps a ContextExecutor with instrumentation.
func WithInstrumentationTx(tx postgres.ContextExecutor) postgres.ContextExecutor {
	return instrumentedTx{ContextExecutor: tx}
}

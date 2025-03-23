package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTraceID(t *testing.T) {
	staticTraceID := "cafebabec0decafe0000000000000000"

	traceID := NewTraceID(t, staticTraceID)
	require.Equal(t, staticTraceID, traceID.String())
}

func TestNewSpanID(t *testing.T) {
	staticSpanID := "deadbeefdeadbeef"

	spanID := NewSpanID(t, staticSpanID)
	require.Equal(t, staticSpanID, spanID.String())
}

func TestNewTraceState(t *testing.T) {
	traceState := "ot=p:8;r:62"

	ts := NewTraceState(t, traceState)
	require.Equal(t, traceState, ts.String())
}

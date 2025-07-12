package iam

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/viebiz/lit/monitoring"
)

func newBufferLogger(t require.TestingT) (*monitoring.Monitor, *bytes.Buffer) {
	logBuffer := bytes.NewBuffer(nil)
	m, err := monitoring.New(monitoring.Config{ServerName: "lightning", Environment: "dev", Version: "1.0.0", Writer: logBuffer})
	require.NoError(t, err)

	return m, logBuffer
}

func TestEnforcerLogger_IsEnabled(t *testing.T) {
	EnforcerLog = false
	mon, _ := newBufferLogger(t)
	logger := enforcerLogger{mon}
	require.False(t, logger.IsEnabled())

	EnforcerLog = true
	require.True(t, logger.IsEnabled())
}

func TestEnforcerLogger_LogModel(t *testing.T) {
	mon, buf := newBufferLogger(t)
	logger := enforcerLogger{mon}

	model := [][]string{{"a", "b"}, {"c"}}

	// WHEN
	logger.LogModel(model)

	// THEN
	out := buf.String()
	require.Contains(t, out, "Model [[a b] [c]]")
}

func TestEnforcerLogger_LogEnforce(t *testing.T) {
	mon, buf := newBufferLogger(t)
	logger := enforcerLogger{mon}
	req := []interface{}{"r1", 2}
	matcher := "M"
	explains := [][]string{{"p"}}

	for _, tc := range []struct {
		enabled bool
		expect  []string
	}{
		{false, nil},
		{true, []string{
			"Enforced false",
			"[r1 2]",
			"hit policy: [[p]]",
		}},
	} {
		// GIVEN
		EnforcerLog = tc.enabled
		buf.Reset()

		// WHEN
		logger.LogEnforce(matcher, req, false, explains)

		// THEN
		out := buf.String()
		if tc.enabled {
			for _, substr := range tc.expect {
				require.Contains(t, out, substr)
			}
		} else {
			require.Empty(t, out)
		}
	}
}

func TestEnforcerLogger_LogRole(t *testing.T) {
	mon, buf := newBufferLogger(t)
	logger := enforcerLogger{mon}

	for _, tc := range []struct {
		enabled bool
		roles   []string
		expect  string
	}{
		{false, []string{"x", "y"}, ""},
		{true, []string{"x", "y"}, "Roles: [x y]"},
	} {
		EnforcerLog = tc.enabled
		buf.Reset()

		logger.LogRole(tc.roles)

		out := buf.String()
		if tc.enabled {
			require.Contains(t, out, tc.expect)
		} else {
			require.Empty(t, out)
		}
	}
}

func TestEnforcerLogger_LogPolicy(t *testing.T) {
	mon, buf := newBufferLogger(t)
	logger := enforcerLogger{mon}

	policy := map[string][][]string{
		"p": {{"a", "b"}},
	}

	for _, tc := range []struct {
		enabled bool
		expect  string
	}{
		{false, ""},
		{true, "Policies: map[p:[[a b]]]"},
	} {
		EnforcerLog = tc.enabled
		buf.Reset()

		logger.LogPolicy(policy)

		out := buf.String()
		if tc.enabled {
			require.Contains(t, out, tc.expect)
		} else {
			require.Empty(t, out)
		}
	}
}

func TestEnforcerLogger_LogError(t *testing.T) {
	mon, buf := newBufferLogger(t)
	logger := enforcerLogger{mon}

	err := errors.New("boom")
	msgs := []string{"first", "second"}

	for _, tc := range []struct {
		enabled bool
		expect  string
	}{
		{false, ""},
		{true, "first,second"},
	} {
		EnforcerLog = tc.enabled
		buf.Reset()

		logger.LogError(err, msgs...)

		out := buf.String()
		if tc.enabled {
			require.True(t, strings.Contains(out, tc.expect), "got %q, want substring %q", out, tc.expect)
			require.True(t, strings.Contains(out, "Err: boom"), "should log the error itself")
		} else {
			require.Empty(t, out)
		}
	}
}

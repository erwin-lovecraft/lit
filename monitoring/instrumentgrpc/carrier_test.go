package instrumentgrpc

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestMdCarrier_Get(t *testing.T) {
	tests := map[string]struct {
		md       metadata.MD
		key      string
		expected string
	}{
		"existing key returns first value": {
			md:       metadata.MD{"traceparent": []string{"value1", "value2"}},
			key:      "traceparent",
			expected: "value1",
		},
		"non-existent key returns empty string": {
			md:       metadata.MD{"traceparent": []string{"value1"}},
			key:      "nonexistent",
			expected: "",
		},
		"empty metadata returns empty string": {
			md:       metadata.MD{},
			key:      "traceparent",
			expected: "",
		},
		"empty value slice returns empty string": {
			md:       metadata.MD{"traceparent": []string{}},
			key:      "traceparent",
			expected: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			carrier := mdCarrier(tt.md)

			// WHEN
			actual := carrier.Get(tt.key)

			// THEN
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestMdCarrier_Set(t *testing.T) {
	tests := map[string]struct {
		md       metadata.MD
		key      string
		value    string
		expected metadata.MD
	}{
		"add value to non-existent key": {
			md:       metadata.MD{},
			key:      "traceparent",
			value:    "value1",
			expected: metadata.MD{"traceparent": []string{"value1"}},
		},
		"add value to existing key": {
			md:       metadata.MD{"traceparent": []string{"value1"}},
			key:      "traceparent",
			value:    "value2",
			expected: metadata.MD{"traceparent": []string{"value1", "value2"}},
		},
		"keys are lowercased": {
			md:       metadata.MD{},
			key:      "TraceParent",
			value:    "value1",
			expected: metadata.MD{"traceparent": []string{"value1"}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			carrier := mdCarrier(tt.md)

			// WHEN
			carrier.Set(tt.key, tt.value)

			// THEN
			require.Equal(t, tt.expected, metadata.MD(carrier))
		})
	}
}

func TestMdCarrier_Keys(t *testing.T) {
	tests := map[string]struct {
		md       metadata.MD
		expected []string
	}{
		"empty metadata returns empty slice": {
			md:       metadata.MD{},
			expected: []string{},
		},
		"single key returns slice with one element": {
			md:       metadata.MD{"traceparent": []string{"value1"}},
			expected: []string{"traceparent"},
		},
		"multiple keys returns all keys": {
			md: metadata.MD{
				"traceparent": []string{"value1"},
				"tracestate":  []string{"value2"},
				"baggage":     []string{"value3"},
			},
			expected: []string{"traceparent", "tracestate", "baggage"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			carrier := mdCarrier(tt.md)

			// WHEN
			actual := carrier.Keys()

			// THEN
			sortedActual := sort.StringSlice(actual)
			sortedActual.Sort()

			sortedExpected := sort.StringSlice(tt.expected)
			sortedExpected.Sort()

			require.Equal(t, sortedExpected, sortedActual)
		})
	}
}

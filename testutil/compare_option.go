package testutil

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Option represents a comparer option that applies to a specific type T
type Option[T any] func() cmp.Option

// IgnoreMapEntries returns an Option that causes the comparison to skip
// map entries that satisfy the given predicate function
func IgnoreMapEntries[K comparable, V any](predicate func(K, V) bool) Option[map[K]V] {
	return func() cmp.Option {
		return cmpopts.IgnoreMapEntries(predicate)
	}
}

// IgnoreSliceMapEntries returns an Option that causes the comparison to skip
// map entries that satisfy the given predicate function
func IgnoreSliceMapEntries[K comparable, V any](predicate func(K, V) bool) Option[[]map[K]V] {
	return func() cmp.Option {
		return cmpopts.IgnoreMapEntries(predicate)
	}
}

func IgnoreSliceElements[T any](predicate func(T) bool) Option[[]T] {
	return func() cmp.Option {
		return cmpopts.IgnoreSliceElements(predicate)
	}
}

func IgnoreUnexported[T any](typs ...any) Option[T] {
	return func() cmp.Option {
		return cmpopts.IgnoreUnexported(typs...)
	}
}

func EquateComparable[T any](typs ...any) Option[T] {
	return func() cmp.Option {
		return cmpopts.EquateComparable(typs...)
	}
}

func SortSlices[S []T, T any](cmpFunc func(T, T) int) Option[S] {
	return func() cmp.Option {
		return cmpopts.SortSlices(cmpFunc)
	}
}

func IgnoreSliceOrder[T any]() Option[[]T] {
	return func() cmp.Option {
		return cmpopts.SortSlices(func(a, b T) int {
			va := reflect.ValueOf(a)
			vb := reflect.ValueOf(b)

			// Convert to strings for comparison
			aStr := fmt.Sprintf("%v", va.Interface())
			bStr := fmt.Sprintf("%v", vb.Interface())

			return strings.Compare(aStr, bStr)
		})
	}
}

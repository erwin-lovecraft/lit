package testutil

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Option represents a comparer option that applies to a specific type T
type Option[T any] interface {
	check(T) // ensures that the option is applied to the correct type at compile-time

	toCmpOption() cmp.Option
}

type ignoreMapEntry[T map[K]V | []map[K]V, K comparable, V any] struct {
	cmp.Option
}

func (e ignoreMapEntry[T, K, V]) check(T) {}

func (e ignoreMapEntry[T, K, V]) toCmpOption() cmp.Option {
	return e.Option
}

// IgnoreMapEntries returns an Option that causes the comparison to skip
// map entries that satisfy the given predicate function
func IgnoreMapEntries[K comparable, V any](predicate func(K, V) bool) Option[map[K]V] {
	return ignoreMapEntry[map[K]V, K, V]{
		Option: cmpopts.IgnoreMapEntries(predicate),
	}
}

// IgnoreSliceMapEntries returns an Option that causes the comparison to skip
// map entries that satisfy the given predicate function
func IgnoreSliceMapEntries[K comparable, V any](predicate func(K, V) bool) Option[[]map[K]V] {
	return ignoreMapEntry[[]map[K]V, K, V]{
		Option: cmpopts.IgnoreMapEntries(predicate),
	}
}

type ignoreSliceElement[T any] struct {
	cmp.Option
}

func (e ignoreSliceElement[T]) check(T) {}

func (e ignoreSliceElement[T]) toCmpOption() cmp.Option {
	return e.Option
}

func IgnoreSliceElements[T any](predicate func(T) bool) Option[[]T] {
	return ignoreSliceElement[[]T]{
		Option: cmpopts.IgnoreSliceElements(predicate),
	}
}

type ignoreUnexported[T any] struct {
	cmp.Option
}

func (e ignoreUnexported[T]) check(T) {}

func (e ignoreUnexported[T]) toCmpOption() cmp.Option {
	return e.Option
}

func IgnoreUnexported[T any](typs ...any) Option[T] {
	return ignoreUnexported[T]{
		Option: cmpopts.IgnoreUnexported(typs...),
	}
}

type equateComparable[T any] struct {
	cmp.Option
}

func (e equateComparable[T]) check(T) {}

func (e equateComparable[T]) toCmpOption() cmp.Option {
	return e.Option
}

func EquateComparable[T any](typs ...any) Option[T] {
	return equateComparable[T]{
		Option: cmpopts.EquateComparable(typs...),
	}
}

type sortSlicesOption[S []T, T any] struct {
	cmp.Option
}

func (e sortSlicesOption[S, T]) check(S) {}

func (e sortSlicesOption[S, T]) toCmpOption() cmp.Option {
	return e.Option
}

func SortSlices[S []T, T any](cmpFunc func(T, T) int) Option[S] {
	return sortSlicesOption[S, T]{
		Option: cmpopts.SortSlices(cmpFunc),
	}
}

func IgnoreSliceOrder[T any]() Option[[]T] {
	return sortSlicesOption[[]T, T]{
		Option: cmpopts.SortSlices(func(a, b T) int {
			va := reflect.ValueOf(a)
			vb := reflect.ValueOf(b)

			// Convert to strings for comparison
			aStr := fmt.Sprintf("%v", va.Interface())
			bStr := fmt.Sprintf("%v", vb.Interface())

			return strings.Compare(aStr, bStr)
		}),
	}
}

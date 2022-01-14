package streams

import (
	"constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

type Stream[T any] interface {
	Next() (item T, ok bool)
}

type onceStream[T any] struct {
	called bool
	generator func() (item T, ok bool)
}

func (s *onceStream[T]) Next() (item T, ok bool) {
	if !s.called {
		s.called = true
		item, ok = s.generator()
	}
	return item, ok
}

func Once[T any](f func() (item T, ok bool)) Stream[T] {
	return &onceStream[T]{generator: f}
}

type generatorStream[T any] struct {
	generator func() (item T, ok bool)
	ended bool
}

func (s *generatorStream[T]) Next() (item T, ok bool) {
	if !s.ended {
		item, ok = s.generator()
		s.ended = !ok
	}
	return item, !s.ended
}

// Generator returns a stream which calls the given function
// with each call to Next() on the returned stream. Once the
// function returns a false ok value, the function will not
// be called again and any successive calls to its Next()
// method will also return a false ok value.
// This may be useful for when you could use FromItems() but
// want to lazily evaluate the results or when you want to
// produce a stream of items that relies on data stored in a
// closure rather than creating a new object which implements
// the Stream[T] interface.
func Generator[T any](f func() (item T, ok bool)) Stream[T] {
	return &generatorStream[T]{generator: f}
}

type finalizerStream[T any] struct {
	Stream[T]
	finalize func()
}

func (s finalizerStream[T]) Next() (item T, ok bool) {
	item, ok = s.Stream.Next()
	if !ok {
		s.finalize()
	}
	return item, ok
}

func WithFinalizer[T any](s Stream[T], finalize func()) Stream[T] {
	return finalizerStream[T]{s, finalize}
}

type initializerStream[T any] struct {
	Stream[T]
	initialize func()
}

func (s *initializerStream[T]) Next() (item T, ok bool) {
	if s.initialize != nil {
		s.initialize()
		s.initialize = nil
	}
	return s.Stream.Next()
}

func WithInitializer[T any](s Stream[T], initialize func()) Stream[T] {
	return &initializerStream[T]{s, initialize}
}

func ForEach[T any](s Stream[T], f func(T)) {
	item, ok := s.Next()
	for ok {
		f(item)
		item, ok = s.Next()
	}
}

type nullStream[T any] struct{}

func (s nullStream[T]) Next() (item T, ok bool) {
	return item, false
}

func Null[T any]() Stream[T] {
	return nullStream[T]{}
}

type sliceStream[T any] struct {
	items []T
	index int
}

func (s *sliceStream[T]) Next() (item T, ok bool) {
	if ok = s.index < len(s.items); ok {
		item = s.items[s.index]
		s.index++
	}
	return item, ok
}

func FromItems[T any](items ...T) Stream[T] {
	return &sliceStream[T]{items: items}
}

func ToSlice[T any](s Stream[T]) (items []T) {
	item, ok := s.Next()
	for ok {
		items = append(items, item)
		item, ok = s.Next()
	}
	return items
}

type counterStream[T Number] struct {
	min, max, step, at T
}

func (s *counterStream[T]) Next() (item T, ok bool) {
	item = s.at
	ok = item < s.max
	if s.step < 0 {
		ok = item > s.min
	}
	if ok {
		s.at += s.step	
	}
	return item, ok
}

func Counter[T Number](min, max, step T) Stream[T] {
	at := min
	if step < 0 {
		at = max
	}
	return &counterStream[T]{
		min: min,
		max: max,
		step: step,
		at: at,
	}
}

type mappedStream[T1, T2 any] struct {
	in Stream[T1]
	mapper func(T1) T2
}

func (s *mappedStream[T1, T2]) Next() (item T2, ok bool) {
	var inItem T1
	if inItem, ok = s.in.Next(); ok {
		item = s.mapper(inItem)
	}
	return item, ok
}

func Map[T1, T2 any](in Stream[T1], mapper func(T1) T2) Stream[T2] {
	return &mappedStream[T1, T2]{
		in: in,
		mapper: mapper,
	}
}

type flatMappedStream[T1, T2 comparable] struct {
	in Stream[T1]
	out Stream[T2]
	mapper func(T1) Stream[T2]
}

func (s *flatMappedStream[T1, T2]) Next() (item T2, ok bool) {
	item, ok = s.out.Next()
	for !ok {
		var inItem T1
		inItem, ok = s.in.Next()
		if !ok {
			return item, false
		}
		s.out = s.mapper(inItem)
		item, ok = s.out.Next()
	}
	return item, true
}

func FlatMap[T1, T2 comparable](in Stream[T1], mapper func(T1) Stream[T2]) Stream[T2] {
	return &flatMappedStream[T1, T2]{
		in: in,
		out: Null[T2](),
		mapper: mapper,
	}
}

type filteredStream[T any] struct {
	in Stream[T]
	filterFunc func(T) bool
}

func (s *filteredStream[T]) Next() (item T, ok bool) {
	for !ok {
		if item, ok = s.in.Next(); !ok {
			return item, false
		}
		ok = s.filterFunc(item)
	}
	return item, true
}

func Filter[T any](in Stream[T], filterFunc func(T) bool) Stream[T] {
	return &filteredStream[T]{
		in: in,
		filterFunc: filterFunc,
	}
}

func Reduce[T any](s Stream[T], init T, reducer func(T, T) T) T {
	reduced := init
	item, ok := s.Next()
	for ok {
		reduced = reducer(reduced, item)
		item, ok = s.Next()
	}
	return reduced
}

func MinFunc[T any](s Stream[T], less func(a, b T) bool) (min T) {
	first, ok := s.Next()
	if !ok {
		return min
	}
	return Reduce(s, first, func(a, b T) T {
		if less(a, b) {
			return a
		}
		return b
	})
}

func Min[T constraints.Ordered](s Stream[T]) T {
	return MinFunc(s, func(a, b T) bool { return a < b })
}

func MaxFunc[T any](s Stream[T], less func(a, b T) bool) (max T) {
	first, ok := s.Next()
	if !ok {
		return max
	}
	return Reduce(s, first, func(a, b T) T {
		if less(a, b) {
			return b
		}
		return a
	})
}

func Max[T constraints.Ordered](s Stream[T]) T {
	return MaxFunc(s, func(a, b T) bool { return a < b })
}

func SumFunc[T any, N Number](s Stream[T], getNum func(T) N) (sum N) {
	return Reduce(Map(s, getNum), sum, func(a, b N) N { return a + b })
}

func Sum[T Number](s Stream[T]) (sum T) {
	return Reduce(s, sum, func(a, b T) T { return a + b })
}

func All[T any](s Stream[T], predicate func(T) bool) bool {
	item, ok := s.Next()
	for ok {
		if !predicate(item) {
			return false
		}
		item, ok = s.Next()
	}
	return true
}

func Any[T any](s Stream[T], predicate func(T) bool) bool {
	item, ok := s.Next()
	for ok {
		if predicate(item) {
			return true
		}
		item, ok = s.Next()
	}
	return false
}


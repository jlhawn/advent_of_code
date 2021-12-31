package slices

import (
	"constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

func ForEach[T any](f func(T), vals ...T) {
	for _, val := range vals {
		f(val)
	}
}

func Reverse[T any](vals []T) {
	for i, j := 0, len(vals)-1; i < j; i, j = i+1, j-1 {
        vals[i], vals[j] = vals[j], vals[i]
    }
}

func Map[T1, T2 any](mapper func(T1) T2, vals ...T1) []T2 {
	mapped := make([]T2, len(vals))
	for i, item := range vals {
		mapped[i] = mapper(item)
	}
	return mapped
}

func Reduce[T any](reducer func(T, T) T, init T, vals ...T) T {
	reduced := init
	for _, val := range vals {
		reduced = reducer(reduced, val)
	}
	return reduced
}

func FlatMap[T1, T2 any](mapper func(T1) []T2, vals...T1) []T2 {
	var flatMapped []T2
	for _, item := range vals {
		flatMapped = append(flatMapped, mapper(item)...)
	}
	return flatMapped
}

func Filter[T any](filter func(T) bool, vals ...T) []T {
	filtered := make([]T, 0, len(vals))
	for _, val := range vals {
		if filter(val) {
			filtered = append(filtered, val)
		}
	}
	return filtered
}

func MinFunc[T any](less func(a, b T) bool, vals ...T) (min T) {
	if len(vals) == 0 {
		return min
	}
	return Reduce(func(a, b T) T {
		if less(a, b) {
			return a
		}
		return b
	}, vals[0], vals[1:]...)
}

func Min[T constraints.Ordered](vals ...T) T {
	return MinFunc(func(a, b T) bool { return a < b }, vals...)
}

func MaxFunc[T any](less func(a, b T) bool, vals ...T) (max T) {
	if len(vals) == 0 {
		return max
	}
	return Reduce(func(a, b T) T {
		if less(a, b) {
			return b
		}
		return a
	}, vals[0], vals[1:]...)
}

func Max[T constraints.Ordered](vals ...T) T {
	return MaxFunc(func(a, b T) bool { return a < b }, vals...)
}

func SumFunc[T any, N Number](getNum func(T) N, vals ...T) (sum N) {
	return Reduce(func(a, b N) N { return a + b }, sum, Map(getNum, vals...)...)
}

func Sum[T Number](vals ...T) (sum T) {
	return Reduce(func(a, b T) T { return a + b }, sum, vals...)
}


package iter

import (
	"errors"
	"sync"
)

type Mapper[T, R any] Iterator[T]

func Map[T, R any](input []T, f func(*T) R) []R {
	return Mapper[T, R]{}.Map(input, f)
}

func (m Mapper[T, R]) Map(input []T, f func(*T) R) []R {
	res := make([]R, len(input))
	Iterator[T](m).ForEachIdx(input, func(i int, t *T) {
		res[i] = f(t)
	})
	return res
}

func MapErr[T, R any](input []T, f func(*T) (R, error)) ([]R, error) {
	return Mapper[T, R]{}.MapErr(input, f)
}

func (m Mapper[T, R]) MapErr(input []T, f func(*T) (R, error)) ([]R, error) {
	var (
		res    = make([]R, len(input))
		errMux sync.Mutex
		errs   []error
	)
	Iterator[T](m).ForEachIdx(input, func(i int, t *T) {
		var err error
		res[i], err = f(t)
		if err != nil {
			errMux.Lock()
			errs = append(errs, err)
			errMux.Unlock()
		}
	})
	return res, errors.Join(errs...)
}

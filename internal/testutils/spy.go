package testutils

import "net/http"

type Spy[T any, R any] struct {
	Calls   []T
	Results []R // outputs
	Target  func(T) R
}

// SpyOn creates a new Spy for functions with a single input and a single output.
//
// Because of Go's generic constraints, this spy supports only functions of the
// form: func(T) R
//
// Functions with multiple arguments must wrap their inputs in a struct or tuple
// Similarly, multiple return values must be wrapped (e.g., using a custom struct or tuples).
//
// This limitation *should* be acceptable for our use case, as we primarily want
// to spy on request-based functions such as `DoSigned` (Node SDK parity).
func SpyOn[T any, R any](target func(T) R) *Spy[T, R] {
	return &Spy[T, R]{Target: target}
}

func (s *Spy[T, R]) Func() func(T) R {
	return func(arg T) R {
		s.Calls = append(s.Calls, arg)

		result := s.Target(arg)
		s.Results = append(s.Results, result)

		return result
	}
}

func (s *Spy[T, R]) CallCount() int {
	return len(s.Calls)
}

func (s *Spy[T, R]) ResultCount() int {
	return len(s.Results)
}

// DoSignedResult is specific for the DoSigned method.
//
// Go does not support multi-value returns for generics (i.e [T, (R, err)]
// Consider adding a generic tuple in the future to avoid defining responses
// for every function that we want to spy on.
//
// Source: https://github.com/golang/go/issues/61920#issuecomment-1676117645
type DoSignedResult struct {
	Response *http.Response
	Error    error
}

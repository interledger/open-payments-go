package testutils

type Spy[T any, R any] struct {
	Calls     []T
	Target    func(T) R
	CallCount int
}

func SpyOn[T any, R any](target func(T) R) *Spy[T, R] {
	return &Spy[T, R]{Target: target}
}

func (s *Spy[T, R]) Func() func(T) R {
	return func(arg T) R {
		s.Calls = append(s.Calls, arg)
		s.CallCount++
		return s.Target(arg)
	}
}

package testutils

type Spy[T any, R any] struct {
	Calls   []T
	Results []R // outputs
	Target  func(T) R
}

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

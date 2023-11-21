package parser

type stream[T any] struct {
	data    []T
	current int
}

func newStream[T any](data []T) *stream[T] {
	return &stream[T]{
		data: data,
	}
}

func (s *stream[T]) isEmpty() bool {
	return s.current >= len(s.data)
}

func (s *stream[T]) prev() T {
	if s.current == 0 {
		return *new(T)
	}

	return s.data[s.current-1]
}

func (s *stream[T]) peek() T {
	if s.isEmpty() {
		return *new(T)
	}

	return s.data[s.current]
}

func (s *stream[T]) peekN(n int) []T {
	if s.isEmpty() {
		return nil
	}

	if s.current+n >= len(s.data) {
		return s.data[s.current:]
	}

	return s.data[s.current : s.current+n]
}

func (s *stream[T]) peekUntil(match func(T) bool) []T {
	if s.isEmpty() {
		return nil
	}

	var sequence []T
	for i := s.current; i < len(s.data); i++ {
		r := s.data[i]
		sequence = append(sequence, r)

		if match(r) {
			break
		}
	}

	return sequence
}

func (s *stream[T]) take() T {
	t := s.takeN(1)
	if len(t) == 0 {
		return *new(T)
	}

	return t[0]
}

func (s *stream[T]) takeN(n int) []T {
	if s.isEmpty() {
		return nil
	}

	got := s.peekN(n)
	s.current += n

	return got
}

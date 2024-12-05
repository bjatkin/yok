package slice

type Itter[T any] struct {
	ready  bool
	offset int
	data   []T
}

func NewIttr[T any](data []T) Itter[T] {
	return Itter[T]{
		data: data,
	}
}

func (i *Itter[T]) Continue() bool {
	return i.offset < len(i.data)
}

func (i *Itter[T]) Next() bool {
	if i.ready {
		i.offset++
	}
	i.ready = true
	return i.offset < len(i.data)
}

func (i *Itter[T]) Item() T {
	if i.offset < 0 || i.offset >= len(i.data) {
		return *new(T)
	}

	return i.data[i.offset]
}

func (i *Itter[T]) Pop(n int) []T {
	first := i.offset
	i.offset += n
	if i.offset >= len(i.data) {
		return i.data[first:]
	}

	return i.data[first:i.offset]
}

func (i *Itter[T]) Peek(n int) []T {
	first := i.offset
	last := first + n
	if last >= len(i.data) {
		return i.data[first:]
	}

	return i.data[first:last]
}

func (i *Itter[T]) All() []T {
	if i.offset >= len(i.data) {
		return i.data[len(i.data)-1:]
	}
	return i.data[i.offset:]
}

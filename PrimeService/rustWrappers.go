package main

type Option[T any] struct {
	val *T
}

func (o Option[T]) Unwrap() T {
	if o.val == nil {
		panic("Trying to unwrap an option of None variant")
	}
	return *o.val
}

func (o Option[T]) Expect(msg string) T {
	if o.val == nil {
		panic(msg)
	}
	return *o.val
}

func (o Option[T]) IsNone() bool {
	if o.val == nil {
		return true
	}
	return false
}

func (o Option[T]) IsSome() bool {
	return !o.IsNone()
}

func Some[T any](val T) Option[T] {
	return Option[T]{val: &val}
}

func None[T any]() Option[T] {
	return Option[T]{val: nil}
}

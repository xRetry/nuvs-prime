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

func (o Option[T]) OkOr(err error) Result[T] {
	if o.IsNone() {
		return Err[T](err)
	}
	return Ok(*o.val)
}

func Some[T any](val T) Option[T] {
	return Option[T]{val: &val}
}

func None[T any]() Option[T] {
	return Option[T]{val: nil}
}

type Result[T any] struct {
	val *T
	err error
}

func (r Result[T]) IsOk() bool {
	if r.err == nil {
		return false
	}
	return true
}

func (r Result[T]) IsErr() bool {
	return !r.IsOk()
}

func (r Result[T]) Unwrap() T {
	if r.IsErr() {
		panic(r.err)
	}
	return *r.val
}

func Ok[T any](val T) Result[T] {
	return Result[T]{val: &val, err: nil}
}

func Err[T any](err error) Result[T] {
	return Result[T]{val: nil, err: err}
}

func ToResult[T any](val T, err error) Result[T] {
	return Result[T]{val: &val, err: err}
}

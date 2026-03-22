package consulx

import "sync/atomic"

type Var[T any] struct {
	v   atomic.Value
	def T
}

func NewVar[T any](def T) *Var[T] {
	x := &Var[T]{def: def}
	x.v.Store(def)
	return x
}

func (x *Var[T]) Load() T {
	return x.v.Load().(T)
}

func (x *Var[T]) Store(val T) {
	x.v.Store(val)
}

func (x *Var[T]) Reset() {
	x.v.Store(x.def)
}

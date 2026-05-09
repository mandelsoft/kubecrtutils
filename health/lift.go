package health

import (
	"github.com/mandelsoft/goutils/generics"
)

func LiftToFactory[T any](p Probe) Factory[T] {
	return FactoryFunc[T](func(c T) Probe {
		return p
	})
}

func Convert[T, S any](factory Factory[S]) Factory[T] {
	return FactoryFunc[T](func(c T) Probe {
		return factory.CreateHealthHandler(generics.Cast[S](c))
	})
}

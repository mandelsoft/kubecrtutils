package health

import (
	"net/http"

	"github.com/mandelsoft/kubecrtutils/types"
)

type Definition[T any] interface {
	types.HealthDefinition[T]
}

type CompositionInterface[D, T any] interface {
	AddReadiness(name string, factory Factory[T]) D
	AddLiveness(name string, factory Factory[T]) D
}

type Factory[T any] interface {
	CreateHealthHandler(T) func(req *http.Request) error
}

type FactoryFunc[T any] func(e T) func(req *http.Request) error

func (f FactoryFunc[T]) CreateHealthHandler(e T) func(req *http.Request) error {
	return f(e)
}

type HealthHandlers[D, T any] struct {
	self      D
	readiness map[string]Factory[T]
	liveness  map[string]Factory[T]
}

var (
	_ Definition[any]                = (*HealthHandlers[any, any])(nil)
	_ CompositionInterface[any, any] = (*HealthHandlers[any, any])(nil)
)

func NewHealthHandlers[D, T any](self D) *HealthHandlers[D, T] {
	return &HealthHandlers[D, T]{
		self:      self,
		readiness: make(map[string]Factory[T]),
		liveness:  make(map[string]Factory[T]),
	}
}

func (h *HealthHandlers[D, T]) AddReadiness(name string, factory Factory[T]) D {
	h.readiness[name] = factory
	return h.self
}

func (h *HealthHandlers[D, T]) AddLiveness(name string, factory Factory[T]) D {
	h.liveness[name] = factory
	return h.self
}

func (h *HealthHandlers[D, T]) ApplyHealthChecks(basename string, c T, mgr types.ControllerManager) error {
	m := mgr.GetManager()
	for name, factory := range h.readiness {
		err := m.AddReadyzCheck(basename+"."+name, factory.CreateHealthHandler(c))
		if err != nil {
			return err
		}
	}
	for name, factory := range h.liveness {
		err := m.AddHealthzCheck(basename+"."+name, factory.CreateHealthHandler(c))
		if err != nil {
			return err
		}
	}
	return nil
}

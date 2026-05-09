package health

import (
	"net/http"

	"github.com/mandelsoft/kubecrtutils/types"
)

type Probe = func(req *http.Request) error

type Definition[T any] interface {
	types.HealthDefinition[T]
}

type CompositionInterface[D, T any] interface {
	AddReadiness(name string, factory Factory[T]) D
	AddLiveness(name string, factory Factory[T]) D
}

type Factory[T any] interface {
	CreateHealthHandler(T) Probe
}

type FactoryFunc[T any] func(e T) Probe

func (f FactoryFunc[T]) CreateHealthHandler(e T) Probe {
	return f(e)
}

// HealthHandlers handles the definition of health handler for a type
// S, which is a subtype of T. T is used for requesting
// applying the probes.
type HealthHandlers[D, S, T any] struct {
	self      D
	readiness map[string]Factory[T]
	liveness  map[string]Factory[T]
}

var (
	_ Definition[any]                = (*HealthHandlers[any, any, any])(nil)
	_ CompositionInterface[any, any] = (*HealthHandlers[any, any, any])(nil)
)

func NewHealthHandlers[D, S, T any](self D) *HealthHandlers[D, S, T] {
	return &HealthHandlers[D, S, T]{
		self:      self,
		readiness: make(map[string]Factory[T]),
		liveness:  make(map[string]Factory[T]),
	}
}

func (h *HealthHandlers[D, S, T]) AddReadiness(name string, factory Factory[S]) D {
	h.readiness[name] = Convert[T](factory)
	return h.self
}

func (h *HealthHandlers[D, S, T]) AddLiveness(name string, factory Factory[S]) D {
	h.liveness[name] = Convert[T](factory)
	return h.self
}

func (h *HealthHandlers[D, S, T]) ApplyHealthChecks(basename string, c T, mgr types.ControllerManager) error {
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

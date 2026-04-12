package logic

import (
	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler/factories"
)

// New like New but uses a default option creator for the options and the a ReconcilationLogic.
func New[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any, F ReconcilationLogic[O, S, P, T]](fac F) *factories.ReconcilerFactory[O, S, P, T] {
	return factories.New[O, S, P, T](
		generics.ObjectFor[O],
		fac.CreateSettings,
		func(def *reconciler.BaseRequest[P], r *factories.Reconciler[O, S, P, T]) reconciler.ReconcileRequest[P] {
			return &Request[O, S, P, T]{
				BaseRequest: def,
				Reconciler:  r,
				logic:       fac,
			}
		})
}

// NewWithOptions is like New, but uses the factory instance
// as options. Therefore, the options are not sharable.
func NewWithOptions[S any, P kubecrtutils.ObjectPointer[T], T any, F ReconcilationLogicWithOptions[F, S, P, T]](fac F) *factories.ReconcilerFactory[F, S, P, T] {
	return factories.New[F, S, P, T](
		func() F { return fac },
		fac.CreateSettings,
		func(def *reconciler.BaseRequest[P], r *factories.Reconciler[F, S, P, T]) reconciler.ReconcileRequest[P] {
			return &Request[F, S, P, T]{
				BaseRequest: def,
				Reconciler:  r,
				logic:       fac,
			}
		})
}

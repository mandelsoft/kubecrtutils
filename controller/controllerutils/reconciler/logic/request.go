package logic

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/controller"
	myreconcile "github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler/factories"
)

// --- begin reconcilation logic ---

type ReconcilationLogic[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] interface {
	CreateSettings(ctx context.Context, o O, controller controller.TypedController[P, T]) (S, error)
	Reconcile(request *Request[O, S, P, T]) myreconcile.Problem
	ReconcileDeleting(request *Request[O, S, P, T]) myreconcile.Problem
	ReconcileDeleted(request *Request[O, S, P, T]) myreconcile.Problem
}

// --- end reconcilation logic ---

type ReconcilationLogicWithOptions[F flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] interface {
	ReconcilationLogic[F, S, P, T]
	flagutils.Options
}

// --- begin request ---
type Request[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct {
	*reconciler.BaseRequest[P]
	Reconciler *factories.Reconciler[O, S, P, T]
	logic      ReconcilationLogic[O, S, P, T]
}

// --- end request ---

func (r *Request[O, S, P, T]) Reconcile() myreconcile.Problem {
	return r.logic.Reconcile(r)
}
func (r *Request[O, S, P, T]) ReconcileDeleting() myreconcile.Problem {
	return r.logic.ReconcileDeleting(r)
}
func (r *Request[O, S, P, T]) ReconcileDeleted() myreconcile.Problem {
	return r.logic.ReconcileDeleted(r)
}

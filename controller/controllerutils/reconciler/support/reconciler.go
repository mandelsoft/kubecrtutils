package support

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/controller"
	myreconcile "github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	Reconciler *Reconciler[O, S, P, T]
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

type OwnerHandler = owner.Handler

// --- begin reconciler ---

// Reconciler includes the options and setting field.
type Reconciler[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct {
	logging.Logger
	Controller   controller.TypedController[P, T]
	GroupKind    schema.GroupKind
	FieldManager string
	Finalizer    string
	OwnerHandler
	Options  O
	Settings S

	request RequestFactory[O, S, P, T]
}

// --- end reconciler ---

func (r *Reconciler[O, S, P, T]) Request(def *reconciler.BaseRequest[P]) reconciler.ReconcileRequest[P] {
	return r.request(def, r)
}

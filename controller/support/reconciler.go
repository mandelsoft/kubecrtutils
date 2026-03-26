package support

import (
	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Reconciler[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct {
	logging.Logger
	Controller   controller.TypedController[P, T]
	GroupKind    schema.GroupKind
	FieldManager string
	Finalizer    string
	Options      O
	Settings     S

	request RequestFactory[O, S, P, T]
}

func (r *Reconciler[O, S, P, T]) Request(def *reconciler.BaseRequest[P]) reconciler.ReconcileRequest[P] {
	return r.request(def, r)
}

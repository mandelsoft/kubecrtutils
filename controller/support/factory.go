package support

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/builder"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	crtreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type None struct{}

func (n None) AddFlags(fs *pflag.FlagSet) {}

type SettingsFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] func(ctx context.Context, o O, controller controller.TypedController[P, T]) (S, error)
type RequestFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] func(*reconciler.BaseRequest[P], *Reconciler[O, S, P, T]) reconciler.ReconcileRequest[P]

type Factory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] interface {
	CreateOptions() O
	CreateSettings(ctx context.Context, o O, controller controller.TypedController[P, T]) (S, error)
	CreateRequest(*reconciler.BaseRequest[P], *Reconciler[O, S, P, T]) reconciler.ReconcileRequest[P]
}

////////////////////////////////////////////////////////////////////////////////

type ReconcilerFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct {
	reqfactory  RequestFactory[O, S, P, T]
	attrfactory SettingsFactory[O, S, P, T]

	*flagutils.OptionsRef[O]
}

func New[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any](o func() O, s SettingsFactory[O, S, P, T], req RequestFactory[O, S, P, T]) *ReconcilerFactory[O, S, P, T] {
	return &ReconcilerFactory[O, S, P, T]{req, s, flagutils.NewOptionsRef[O](o)}
}

func NewByFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any](fac Factory[O, S, P, T]) *ReconcilerFactory[O, S, P, T] {
	return New[O, S, P, T](fac.CreateOptions, fac.CreateSettings, fac.CreateRequest)
}

func (f *ReconcilerFactory[O, S, P, T]) CreateReconciler(ctx context.Context, controller controller.TypedController[P, T], b builder.Builder) (crtreconcile.Reconciler, error) {
	var err error
	var set S

	if f.attrfactory != nil {
		set, err = f.attrfactory(ctx, f.Options, controller)
		if err != nil {
			return nil, err
		}
	}

	gvks, _, err := controller.GetCluster().GetScheme().ObjectKinds(controller.GetResource())
	if err != nil {
		return nil, err
	}

	r := &Reconciler[O, S, P, T]{
		Controller:   controller,
		Logger:       controller.GetLogger(),
		FieldManager: controller.GetFieldManager(),
		Finalizer:    controller.GetFinalizer(),
		GroupKind: schema.GroupKind{
			Group: gvks[0].Group,
			Kind:  gvks[0].Kind,
		},
		Options:  f.Options,
		Settings: set,
		request:  f.reqfactory,
	}
	r.Info("using {{ctype}} {{cluster}}[{{info}}]", "ctype", controller.GetCluster().GetTypeInfo(), "name", controller.GetCluster().GetName(), "info", controller.GetCluster().GetInfo())
	return reconciler.CRTReconcilerFor[P](controller, r, 0), nil
}

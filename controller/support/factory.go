package support

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/reflectutils"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/builder"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
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

// ReconcilerFactory creates a controller.Reconciler for given options and settings type.
// The options can implement ModifyFinalizer to influence the
// generated finalizer name.
type ReconcilerFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct {
	reqfactory  RequestFactory[O, S, P, T]
	attrfactory SettingsFactory[O, S, P, T]

	*flagutils.OptionsRef[O]
}

// New provides a factory for a reconciler based on a support factory handling Options
// creating a support reconciler providing the Options and a Settings object created based
// by the given meta factory. The settings can be created based on the final
// controller instance providing access to all elements defclared in the controller definition.
func New[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any](o func() O, s SettingsFactory[O, S, P, T], req RequestFactory[O, S, P, T]) *ReconcilerFactory[O, S, P, T] {
	return &ReconcilerFactory[O, S, P, T]{req, s, flagutils.NewOptionsRef[O](o)}
}

func NewByFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any](fac Factory[O, S, P, T]) *ReconcilerFactory[O, S, P, T] {
	return New[O, S, P, T](fac.CreateOptions, fac.CreateSettings, fac.CreateRequest)
}

func (r *ReconcilerFactory[O, S, P, T]) ModifyFinalizer(f string) string {
	reflectutils.CallOptionalInterfaceMethodOn[controller.FinalizerModifier](r.Options, f)
	if m, ok := generics.TryCast[controller.FinalizerModifier](r.Options); ok {
		return m.ModifyFinalizer(f)
	}
	return f
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
	r.Info("using main {{ctype}} {{cluster}}[{{info}}]", "ctype", controller.GetCluster().GetTypeInfo(), "cluster", controller.GetCluster().GetName(), "info", controller.GetCluster().GetInfo())
	return reconciler.CRTReconcilerFor[P](controller, r, 0), nil
}

////////////////////////////////////////////////////////////////////////////////

type DefaultFactory[O flagutils.Options, A any, P kubecrtutils.ObjectPointer[T], T any] struct{}

var _ Factory[None, None, *v1.Secret, v1.Secret] = (*testFactory[None, None, *v1.Secret, v1.Secret])(nil)

func (f *DefaultFactory[O, S, P, T]) CreateOptions() O {
	return generics.ObjectFor[O]()
}

func (f *DefaultFactory[O, S, P, T]) CreateSettings(ctx context.Context, o None, controller controller.TypedController[P, T]) (S, error) {
	return generics.ObjectFor[S](), nil
}

type testFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct {
	DefaultFactory[O, S, P, T]
}

func (f *testFactory[O, S, P, T]) CreateRequest(r *reconciler.BaseRequest[P], r2 *Reconciler[None, None, P, T]) reconciler.ReconcileRequest[P] {
	panic("implement me")
}

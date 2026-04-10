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
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/modern-go/reflect2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	crtreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type None = reconciler.None

type SettingsFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] func(ctx context.Context, o O, controller controller.TypedController[P, T]) (S, error)
type RequestFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] func(*reconciler.BaseRequest[P], *Reconciler[O, S, P, T]) reconciler.ReconcileRequest[P]

type Factory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] interface {
	CreateOptions() O
	CreateSettings(ctx context.Context, o O, controller controller.TypedController[P, T]) (S, error)
	CreateRequest(*reconciler.BaseRequest[P], *Reconciler[O, S, P, T]) reconciler.ReconcileRequest[P]
}

type FactoryWithOptions[F flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] interface {
	flagutils.Options
	CreateSettings(ctx context.Context, o F, controller controller.TypedController[P, T]) (S, error)
	CreateRequest(*reconciler.BaseRequest[P], *Reconciler[F, S, P, T]) reconciler.ReconcileRequest[P]
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
// controller instance providing access to all elements declared in the controller definition.
// The Option type is prepared to be sharable among multiple controllers using a flagutils.NewOptionsRef.
func New[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any](o func() O, s SettingsFactory[O, S, P, T], req RequestFactory[O, S, P, T]) *ReconcilerFactory[O, S, P, T] {
	return &ReconcilerFactory[O, S, P, T]{req, s, flagutils.NewOptionsRef[O](o)}
}

func NewByFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any](fac Factory[O, S, P, T]) *ReconcilerFactory[O, S, P, T] {
	return New[O, S, P, T](fac.CreateOptions, fac.CreateSettings, fac.CreateRequest)
}

// NewByFactoryWithOptions is like NewByFactory, but uses the factory instance
// as options. Therefore, the options are not sharable.
func NewByFactoryWithOptions[S any, P kubecrtutils.ObjectPointer[T], T any, F FactoryWithOptions[F, S, P, T]](fac F) *ReconcilerFactory[F, S, P, T] {
	return New[F, S, P, T](func() F { return fac }, fac.CreateSettings, fac.CreateRequest)
}

// NewByLogicis like New but uses a default option creator for the options and the a ReconcilationLogic.
func NewByLogic[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any, F ReconcilationLogic[O, S, P, T]](fac F) *ReconcilerFactory[O, S, P, T] {
	return New[O, S, P, T](
		generics.ObjectFor[O],
		fac.CreateSettings,
		func(def *reconciler.BaseRequest[P], r *Reconciler[O, S, P, T]) reconciler.ReconcileRequest[P] {
			return &Request[O, S, P, T]{
				BaseRequest: def,
				Reconciler:  r,
				logic:       fac,
			}
		})
}

// NewByLogicWithOptions is like NewByLogic, but uses the factory instance
// as options. Therefore, the options are not sharable.
func NewByLogicWithOptions[S any, P kubecrtutils.ObjectPointer[T], T any, F ReconcilationLogicWithOptions[F, S, P, T]](fac F) *ReconcilerFactory[F, S, P, T] {
	return New[F, S, P, T](
		func() F { return fac },
		fac.CreateSettings,
		func(def *reconciler.BaseRequest[P], r *Reconciler[F, S, P, T]) reconciler.ReconcileRequest[P] {
			return &Request[F, S, P, T]{
				BaseRequest: def,
				Reconciler:  r,
				logic:       fac,
			}
		})
}

func (r *ReconcilerFactory[S, P, T, F]) ModifyFinalizer(f string) string {
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

	var oh owner.Handler
	if !reflect2.IsNil(f.Options) {
		if p, ok := any(f.Options).(owner.HandlerProvider); ok {
			oh = p.GetOwnerHandler(controller.GetCluster())
		}
	}
	if oh == nil {
		oh = owner.NewHandler(controller.GetCluster())
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
		OwnerHandler: oh,
		Options:      f.Options,
		Settings:     set,
		request:      f.reqfactory,
	}
	r.Info("  using main {{ctype}} {{cluster}}[{{info}}]", "ctype", controller.GetCluster().GetTypeInfo(), "cluster", controller.GetCluster().GetName(), "info", controller.GetCluster().GetInfo())
	return reconciler.CRTReconcilerFor[P](controller, r, 0), nil
}

////////////////////////////////////////////////////////////////////////////////

type DefaultFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct{}

var _ Factory[None, None, *v1.Secret, v1.Secret] = (*testFactory[None, None, *v1.Secret, v1.Secret])(nil)

func (f *DefaultFactory[O, S, P, T]) CreateOptions() O {
	return generics.ObjectFor[O]()
}

func (f *DefaultFactory[O, S, P, T]) CreateSettings(ctx context.Context, o O, controller controller.TypedController[P, T]) (S, error) {
	return generics.ObjectFor[S](), nil
}

type testFactory[O flagutils.Options, S any, P kubecrtutils.ObjectPointer[T], T any] struct {
	DefaultFactory[O, S, P, T]
}

func (f *testFactory[O, S, P, T]) CreateRequest(r *reconciler.BaseRequest[P], r2 *Reconciler[O, S, P, T]) reconciler.ReconcileRequest[P] {
	panic("implement me")
}

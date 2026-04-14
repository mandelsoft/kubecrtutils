package reconciler

import (
	"context"
	"reflect"
	"time"

	"github.com/go-test/deep"
	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/builder"
	myreconcile "github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	builder2 "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type CRTReconciler interface {
	reconcile.Reconciler
	GetEffective() any
}

// decrepated: use RequestFactory
type Reconciler[T client.Object] interface {
	RequestFactory[T]
}

type RequestFactory[T client.Object] interface {
	Request(def *BaseRequest[T]) ReconcileRequest[T]
}

// --- begin reconcilation logic ---

type ReconcilationLogic interface {
	Reconcile() myreconcile.Problem
	ReconcileDeleting() myreconcile.Problem
	ReconcileDeleted() myreconcile.Problem
}

// --- end reconcilation logic ---

// --- begin request ----

type Request[T client.Object] interface {
	context.Context
	logging.Logger
	cluster.Cluster
	record.EventRecorder

	GetOptions() flagutils.Options

	GetKey() client.ObjectKey
	GetObject() T
	GetOrig() T

	StatusChanged() bool
	UpdateStatus() myreconcile.Problem
	TriggerStatusChanged()
	GetAfter() time.Duration
}

// --- end request ---

// --- begin reconcile request ----

type ReconcileRequest[T client.Object] interface {
	Request[T]
	ReconcilationLogic
}

// --- end reconcile request ----

type BaseRequest[T client.Object] struct {
	cluster.Cluster
	record.EventRecorder
	context.Context
	logging.Logger
	types.OwnerHandler
	controller types.Controller
	// --- begin request fields ---
	mcreconcile.Request
	Object T
	Orig   T
	After  time.Duration
	// --- end request fields ---
}

func (r *BaseRequest[T]) GenerateNameFor(tgt, prefix string, len ...int) string {
	return r.controller.GenerateNameFor(r.Context, r.controller.GetClusters().Get(tgt).AsCluster(), prefix, r.Object.GetNamespace(), r.Object.GetName())
}

func (r *BaseRequest[T]) GetController() types.Controller {
	return r.controller
}

func (r *BaseRequest[T]) GetScheme() *runtime.Scheme {
	return r.Cluster.GetScheme()
}

func (r *BaseRequest[T]) GetOptions() flagutils.Options {
	return r.controller.GetOptions()
}

func (r *BaseRequest[T]) GetObject() T {
	return r.Object
}

func (r *BaseRequest[T]) GetOrig() T {
	return r.Orig
}

func (r *BaseRequest[T]) GetKey() client.ObjectKey {
	return r.Request.NamespacedName
}

func (r *BaseRequest[T]) GetAfter() time.Duration {
	return r.After
}

func (r *BaseRequest[T]) GetLogicalCluster(name string) types.ClusterEquivalent {
	return r.GetController().GetLogicalCluster(name)
}

func (r *BaseRequest[T]) StatusChanged() bool {
	v := reflect.ValueOf(r.Object).Elem().FieldByName("Status")
	if !v.IsValid() {
		return false
	}

	n := v.Interface()
	o := reflect.ValueOf(r.Orig).Elem().FieldByName("Status").Interface()
	diff := deep.Equal(n, o)
	if len(diff) > 0 {
		r.Info("status changed", "diff", diff)
		return true
	}
	return false
}

func (r *BaseRequest[T]) UpdateStatus() myreconcile.Problem {
	err := r.Cluster.Status().Update(r, r.Object)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return myreconcile.TemporaryProblem(err)
	}
	return nil
}

func (r *BaseRequest[T]) TriggerStatusChanged() {
}

// SetStatusCondition assumes there is a field Status.Conditions of type []metav1.Condition.
func (r *BaseRequest[T]) SetStatusCondition(obj T, condition metav1.Condition) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	s := v.FieldByName("Status")

	if !s.IsValid() {
		return false
	}
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	c := s.FieldByName("Conditions")
	if !c.IsValid() {
		return false
	}

	if condition.ObservedGeneration == 0 {
		condition.ObservedGeneration = obj.GetGeneration()
	}
	return meta.SetStatusCondition(c.Addr().Interface().(*[]metav1.Condition), condition)
}

// Get help Goland resolve this method from interface cluster.Cluster.
func (r *BaseRequest[T]) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return r.Cluster.Get(ctx, key, obj, opts...)
}

type DefaultReconcileRequest[T client.Object, R any] struct {
	BaseRequest[T]
	Reconciler R
}

var _ ReconcileRequest[client.Object] = (*DefaultReconcileRequest[client.Object, any])(nil)

func (r *DefaultReconcileRequest[T, R]) ReconcileDeleted() myreconcile.Problem {
	return nil
}

func (r *DefaultReconcileRequest[T, R]) ReconcileDeleting() myreconcile.Problem {
	return nil
}

func (r *DefaultReconcileRequest[T, R]) Reconcile() myreconcile.Problem {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type None struct{}

func (n None) AddFlags(fs *pflag.FlagSet) {}

type ReconciliationByReconcileFunc[P kubecrtutils.ObjectPointer[T], T any] func(request ReconcileRequest[P]) myreconcile.Problem

func (f ReconciliationByReconcileFunc[P, T]) CreateReconciler(ctx context.Context, controller controller.TypedController[P, T], b *builder2.Builder) (reconcile.Reconciler, error) {
	return CRTReconcilerFor[P, T](controller, &defaultReconciler[P, T]{f}, 0), nil
}

type defaultReconciler[P kubecrtutils.ObjectPointer[T], T any] struct {
	reconcile ReconciliationByReconcileFunc[P, T]
}

func (r *defaultReconciler[P, T]) Request(base *BaseRequest[P]) ReconcileRequest[P] {
	return &DefaultReconcileRequest[P, *defaultReconciler[P, T]]{
		BaseRequest: *base,
		Reconciler:  r,
	}
}

////////////////////////////////////////////////////////////////////////////////

type DefaultRequestFactoryFuncWithOptions[O flagutils.Options, P kubecrtutils.ObjectPointer[T], T any] = func(opts O, def *BaseRequest[P]) ReconcileRequest[P]
type DefaultRequestFactoryFunc[P kubecrtutils.ObjectPointer[T], T any] = DefaultRequestFactoryFuncWithOptions[None, P, T]

type DefaultFactory[O flagutils.Options, P kubecrtutils.ObjectPointer[T], T any] struct {
	flagutils.OptionsRef[O]
	factory DefaultRequestFactoryFuncWithOptions[O, P, T]
}

var (
	_ controller.ReconcilerFactory[*corev1.Secret, corev1.Secret] = (*DefaultFactory[flagutils.Options, *corev1.Secret, corev1.Secret])(nil)
	_ flagutils.Options                                           = (*DefaultFactory[flagutils.Options, *corev1.Secret, corev1.Secret])(nil)
	_ flagutils.Preparable                                        = (*DefaultFactory[flagutils.Options, *corev1.Secret, corev1.Secret])(nil)
)

func New[P kubecrtutils.ObjectPointer[T], T any](rf DefaultRequestFactoryFunc[P, T]) controller.ReconcilerFactory[P, T] {
	return NewWithOptions[None, P, T](rf)
}

// NewWithOptions provides a factory for standard reconcile.Reconciler using
// a DefaultRequestFactoryFuncWithOptions to create high level ReconcileRequest objects for the execution
// of reconcilation requests. The given flagutils.Options type is used to extend the option handling.
func NewWithOptions[O flagutils.Options, P kubecrtutils.ObjectPointer[T], T any](rf DefaultRequestFactoryFuncWithOptions[O, P, T]) controller.ReconcilerFactory[P, T] {
	f := &DefaultFactory[O, P, T]{
		OptionsRef: *flagutils.NewOptionsRef[O](generics.ObjectFor[O]),
		factory:    rf,
	}
	return f
}

func (d *DefaultFactory[O, P, T]) CreateReconciler(ctx context.Context, c controller.TypedController[P, T], b builder.Builder) (reconcile.Reconciler, error) {
	return CRTReconcilerFor[P, T](c, d), nil
}

func (d *DefaultFactory[O, P, T]) Request(def *BaseRequest[P]) ReconcileRequest[P] {
	return d.factory(d.Options, def)
}

func (d *DefaultFactory[O, P, T]) AddFlags(fs *pflag.FlagSet) {
	d.Options.AddFlags(fs)
}

////////////////////////////////////////////////////////////////////////////////

type crtReconciler[T client.Object] struct {
	name       string
	controller types.Controller
	cluster    types.ClusterEquivalent
	logger     logging.Logger
	reconciler RequestFactory[T]
	after      time.Duration
}

var _ CRTReconciler = (*crtReconciler[client.Object])(nil)

func CRTReconcilerFor[P kubecrtutils.ObjectPointer[T], T any](c controller.TypedController[P, T], r RequestFactory[P], after ...time.Duration) CRTReconciler {
	return &crtReconciler[P]{
		name:       c.GetName(),
		logger:     c.GetLogger(),
		cluster:    c.GetCluster(),
		controller: c,
		reconciler: r,
		after:      general.Optional(after...),
	}
}

func (d *crtReconciler[T]) GetEffective() any {
	return d.reconciler
}

func (d *crtReconciler[T]) GetController() types.Controller {
	return d.controller
}

func (d *crtReconciler[T]) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	var _nil T
	var prob myreconcile.Problem

	var cl cluster.Cluster
	var clusterName string
	if d.cluster.AsCluster() != nil {
		cl = d.cluster.AsCluster()
		clusterName = cl.GetName()
	} else {
		cl = clustercontext.ClusterFor(ctx)
		clusterName = clustercontext.ClusterNameFor(ctx)
	}

	cl.WaitForCacheSync(ctx)

	req := BaseRequest[T]{
		controller:    d.controller,
		Context:       ctx,
		Cluster:       cl,
		EventRecorder: cl.GetEventRecorderFor(d.name),
		OwnerHandler:  d.controller.GetOwnerHandler(),
		Logger:        d.logger.WithName(request.String()).WithValues("object", request.NamespacedName, "cluster", cl.GetName(), "effcluster", cl.GetEffective().GetName()),
		Request:       mcreconcile.Request{request, clusterName},
		After:         d.after,
	}

	req.Orig = reflect.New(reflect.TypeFor[T]().Elem()).Interface().(T)
	var effreq ReconcileRequest[T]

	err := req.Cluster.Get(ctx, request.NamespacedName, req.Orig)
	if err != nil {
		if errors.IsNotFound(err) {
			req.Orig = _nil
			effreq = d.reconciler.Request(&req)
			effreq.Info("*** reconcile deleted")
			prob = effreq.ReconcileDeleted()
			return myreconcile.Result(effreq, prob)
		} else {
			return reconcile.Result{}, err
		}
	} else {
		req.Object = req.Orig.DeepCopyObject().(T)
		effreq = d.reconciler.Request(&req)
		if effreq.GetObject().GetDeletionTimestamp().IsZero() {
			effreq.Info("*** reconcile")
			prob = effreq.Reconcile()
		} else {
			effreq.Info("*** reconcile deletion")
			prob = effreq.ReconcileDeleting()
		}

	}
	if effreq.StatusChanged() {
		prob = myreconcile.AggregateProblem(prob, effreq.UpdateStatus())
		if prob == nil {
			effreq.TriggerStatusChanged()
		}
	}
	return myreconcile.Result(effreq, prob, effreq.GetAfter())
}

package replicate

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconcile"
	"github.com/mandelsoft/kubecrtutils/controller/controllerutils/reconciler/support"
	"github.com/mandelsoft/kubecrtutils/examples/simple/controllers"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// --- begin reconcilation logic ---

type Request = *support.Request[*controllers.Options, *controllers.Settings, *Resource, Resource]

type ReconcilationLogic struct {
}

func (r *ReconcilationLogic) CreateSettings(ctx context.Context, o *controllers.Options, c controller.TypedController[*Resource, Resource]) (*controllers.Settings, error) {
	s := &controllers.Settings{
		Source:  c.GetClusters().Get(controllers.SOURCE),
		Target:  c.GetClusters().Get(controllers.TARGET).AsCluster(),
		Mapping: controllers.NewMapping(),
	}

	log := c.GetLogger()
	log.Info("using source {{type}} {{name}}[{{info}}]", s.Source.GetTypeInfo(), s.Source.GetEffective().GetName(), s.Source.GetTypeInfo())
	log.Info("using target {{type}} {{name}}[{{info}}]", s.Target.GetTypeInfo(), s.Target.GetEffective().GetName(), s.Target.GetTypeInfo())
	return s, nil
}

// --- end reconcilation logic ---

func (l *ReconcilationLogic) Reconcile(r Request) reconcile.Problem {
	hasStatus := false // no status for our resource

	obj := r.GetObject()

	if objutils.GetAnnotation(obj, controllers.REPLICATED_ANNOTATION) != "" {
		r.Info("skip replicated object")
		return nil
	}
	s := r.Reconciler.Settings

	ok := l.IsResponsible(r)

	if !ok {
		r.Info("handle replica deletion for being not responsible")
		p := r.ReconcileDeleting()
		if p != nil {
			return p
		}
		return nil
	}

	// --- begin finalizer ---
	patch := client.MergeFrom(r.GetOrig())
	if controllerutil.AddFinalizer(obj, r.Reconciler.Finalizer) {
		if err := r.Patch(r, r.Object, patch); err != nil {
			return reconcile.TemporaryProblem(client.IgnoreNotFound(err))
		}
		r.Info("taking responsibility")
	}
	// --- end finalizer ---

	// --- begin mapping ---
	// assure target namespace
	namespace := objutils.GenerateUniqueName("replica", r.Cluster.GetId(), "", r.Namespace, objutils.MAX_NAMESPACELEN)
	key := client.ObjectKey{
		Name:      r.Name,
		Namespace: namespace,
	}

	mctx := controllers.WithCluster(r, s.Target)
	prob := s.Mapping.SetOriginal(mctx, key, r.Request)
	if prob != nil {
		return prob
	}
	// --- end mapping ---

	// --- begin prepare ---
	// update replica
	newp := r.Object.DeepCopyObject().(*Resource)
	newp.SetNamespace(namespace)
	objutils.CleanupMeta(newp)
	objutils.SetAnnotation(newp, controllers.REPLICATED_ANNOTATION, r.Cluster.GetId())
	controllerutil.RemoveFinalizer(newp, r.Reconciler.Finalizer)

	err := r.Reconciler.SetOwner(r.Cluster, r.Object, s.Target, newp)
	if err != nil {
		return reconcile.TemporaryProblem(err)
	}
	// --- end prepare ---

	var tgt Resource
	tgtp := &tgt
	err = s.Target.Get(r.Context, key, tgtp)
	if err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.TemporaryProblem(err)
		}
		r.Info("create in target")
		err = s.Target.Create(r.Context, newp, &client.CreateOptions{
			FieldManager: r.Reconciler.FieldManager,
		})
	} else {
		if tgtp.GetDeletionTimestamp() != nil {
			// complete deletion before recreation
			r.Info("replica is deleting -> wait to be completed")
			return nil
		}
		newp.SetFinalizers(tgtp.GetFinalizers())

		if hasStatus {
			// --- begin status ---
			status, err := objutils.GetStatusField(tgtp)
			if err != nil {
				r.Info("cannot determine status field")
				return nil
			}
			err = objutils.SetStatusField(newp, status)
			if err != nil {
				r.Info("cannot set status field")
				return nil
			}
			// --- end status ---
		}

		// pass s.FieldManager to patch only managed fields
		_, err = cluster.ClientSideApplyObject(s.Target, cluster.DefaultOperationContext(r, r, ""), newp, tgtp)
	}
	return reconcile.TemporaryProblem(client.IgnoreNotFound(err))
}

// --- begin deleting ---
func (*ReconcilationLogic) ReconcileDeleting(r Request) reconcile.Problem {
	namespace := objutils.GenerateUniqueName("replica", r.Cluster.GetId(), "", r.Namespace, objutils.MAX_NAMESPACELEN)
	s := r.Reconciler.Settings

	var tgt Resource
	tgtp := &tgt
	key := client.ObjectKey{Name: r.Name, Namespace: namespace}
	err := s.Target.Get(r.Context, key, tgtp)
	if err != nil {
		if !errors.IsNotFound(err) {
			return reconcile.TemporaryProblem(err)
		}
		r.Info("replica already deleted")
		patch := client.MergeFrom(r.GetOrig())
		if controllerutil.RemoveFinalizer(r.Object, r.Reconciler.Finalizer) {
			r.Info("releasing responsibility")
			if err := r.Patch(r, r.Object, patch); err != nil {
				return reconcile.TemporaryProblem(client.IgnoreNotFound(err))
			}
		}
		mctx := controllers.WithCluster(r, s.Target)
		return s.Mapping.RemoveOriginal(mctx, key)
	}
	patch := client.MergeFrom(tgtp.DeepCopyObject().(client.Object))
	if controllerutil.RemoveFinalizer(tgtp, r.Reconciler.Finalizer) {
		r.Info("removing finalizer from replica")
		if err := r.Patch(r, tgtp, patch); err != nil {
			return reconcile.TemporaryProblem(client.IgnoreNotFound(err))
		}
	}

	if tgtp.GetDeletionTimestamp() != nil {
		r.Info("replica already deleting")
	} else {
		r.Info("request replica deletion")
		return reconcile.TemporaryProblem(s.Target.Delete(r.Context, tgtp))
	}
	return reconcile.Succeeded()
}

// --- end deleting ---

func (*ReconcilationLogic) ReconcileDeleted(r Request) reconcile.Problem {
	return reconcile.Succeeded()
}

func (*ReconcilationLogic) IsResponsible(r Request) bool {
	return objutils.CheckAnnotation(r.GetObject(), r.Reconciler.Options.Annotation, r.Reconciler.Options.Class)
}

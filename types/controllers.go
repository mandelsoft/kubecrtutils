package types

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types/plain"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ControllerNames = plain.ControllerNames

type ControllerSet interface {
	GetNames() []string
	GetGroups() map[string][]string
}

type ControllerSource interface {
	GetControllerSet() ControllerSet
}

type ControllerDefinition interface {
	ComponentDefinition

	GetCluster() string
	GetResource() client.Object
	GetGroups() NameSet
	GetWatchPredicates() []predicate.Predicate
	GetFinalizer() string
}

type ControllerDefinitions interface {
	internal.Definitions[ControllerDefinition, ControllerDefinitions]
	AddRule(...Constraint) ControllerDefinitions

	flagutils.Validatable

	ControllerSource
	ClusterFilter
	ComponentFilter

	IndexProvider
	Applyable
}

// --- begin controller ---

type Controller interface {
	Component

	GetMainCluster() ClusterEquivalent
	GetFieldManager() string
	GetFinalizer() string
	GetResource() client.Object
	GetGroupKind() schema.GroupKind
	GetRecoder(ctx context.Context) record.EventRecorder
	GetReconciler() reconcile.Reconciler
	GetOwnerHandler() OwnerHandler

	Complete(ctx context.Context) error

	GenerateNameFor(ctx context.Context, tgt Cluster, prefix, namespace, name string, len ...int) string
}

// --- end controller ---

type Controllers interface {
	internal.Group[Controller]
}

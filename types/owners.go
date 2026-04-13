package types

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// --- begin cluster matcher ---

// ClusterMatcher checks based on a cluster id whether a cluster should be
// considered. If yes the name of the cluster is returned, also.
type ClusterMatcher func(clusterId string) (clusterName string, equal bool)

// --- end cluster matcher ---

// Owner describes an owner of an object.
// including group, kine, namespace, name and cluster
type Owner = TypedGlobalKey

var NewOwner = NewGlobalTypedKey

// --- begin owner handler ---

type OwnerHandler interface {
	SetOwner(cluster Cluster, owner client.Object, target Cluster, slave client.Object) error
	// GetOwner extracts the owner of a dedicated type for obj in cluster target for
	// clusters matched by cmatch.
	GetOwner(cmatch ClusterMatcher, target Cluster, obj client.Object, kind schema.GroupKind) (string, *client.ObjectKey)
	GetOwners(cmatch ClusterMatcher, targetId string, obj client.Object) []Owner
}

// --- end owner handler ---

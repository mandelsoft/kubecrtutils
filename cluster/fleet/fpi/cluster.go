package fpi

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Cluster struct {
	fleet types.Fleet
	types.Cluster
}

func NewCluster(fleet types.Fleet, cluster types.Cluster) *Cluster {
	return &Cluster{
		fleet:   fleet,
		Cluster: cluster,
	}
}

func (c *Cluster) AsCluster() types.Cluster {
	return c
}

func (c *Cluster) EnqueueByGVK(ctx context.Context, gvk schema.GroupVersionKind, key client.ObjectKey) error {
	return c.fleet.EnqueueByGVK(clustercontext.WithCluster(ctx, c), gvk, key)
}

func (c *Cluster) EnqueueByObject(ctx context.Context, obj runtime.Object) error {
	return c.fleet.EnqueueByObject(clustercontext.WithCluster(ctx, c), obj)
}

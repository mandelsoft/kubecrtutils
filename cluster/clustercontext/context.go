package clustercontext

import (
	"context"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/types"
)

var key = "cluster"

func WithCluster(ctx context.Context, cluster types.Cluster) context.Context {
	return context.WithValue(ctx, &key, cluster)
}

func ClusterFor(ctx context.Context) types.Cluster {
	return generics.Cast[types.Cluster](ctx.Value(&key))
}

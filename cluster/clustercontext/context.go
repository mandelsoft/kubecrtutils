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

////////////////////////////////////////////////////////////////////////////////

var nkey = "clustername"

func WithClusterName(ctx context.Context, cluster string) context.Context {
	return context.WithValue(ctx, &nkey, cluster)
}

func ClusterNameFor(ctx context.Context) string {
	return generics.Cast[string](ctx.Value(&nkey))
}

////////////////////////////////////////////////////////////////////////////////

func WithClusterAndName(ctx context.Context, cluster types.Cluster, name string) context.Context {
	return WithCluster(WithClusterName(ctx, name), cluster)
}

package handler

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/types"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
)

func Lift[F any](f F) ControllerAware[ClusterAware[F]] {
	return LiftToController(LiftToCluster(f))
}

func LiftToCluster[F any](f F) ClusterAware[F] {
	return func(clusterName string, cluster sigcluster.Cluster) F {
		return f
	}
}

func LiftToController[F any](f F) ControllerAware[F] {
	return func(ctx context.Context, c types.Controller) (F, error) {
		return f, nil
	}
}

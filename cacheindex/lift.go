package cacheindex

import (
	"context"

	"github.com/mandelsoft/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Lift[T any](f T) ClustersAware[T] {
	return func(ctx context.Context, logger logging.Logger, clusters Clusters) (T, error) {
		return f, nil
	}
}

func ConvertIndexerFunc[T client.Object](f IndexerFunc[T]) client.IndexerFunc {
	return func(object client.Object) []string {
		return f(any(object).(T))
	}
}

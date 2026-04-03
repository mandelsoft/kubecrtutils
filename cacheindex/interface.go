package cacheindex

import (
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Index = types.Index
type Indices = types.Indices

type ClusterEquivalent = types.ClusterEquivalent
type Clusters = types.Clusters

type ClustersAware[T any] = types.ClustersAware[T]

type IndexerFunc[T client.Object] = types.IndexerFunc[T]
type IndexerFactory = types.IndexerFactory

type TypedIndexerFunc[P client.Object] = func(P) []string
type TypedIndexerFactory[P client.Object] = ClustersAware[TypedIndexerFunc[P]]

type Definition = types.IndexDefinition

package cacheindex

import (
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types"
)

type _indices struct {
	internal.Group[types.Index]
}

func NewIndices() Indices {
	return &_indices{internal.NewGroup[types.Index]("index")}
}

package controller

import (
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/types"
)

type Controllers = types.Controllers

type Controller[P kubecrtutils.ObjectPointer[T], T any] interface {
	types.Controller
	GetDefinition() TypedDefinition[P, T]
	GetTypedIndex(name string) cacheindex.TypedIndex[T]
}

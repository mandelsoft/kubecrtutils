package controller

import (
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/types"
)

type Controllers = types.Controllers

type Controller[T any, P kubecrtutils.ObjectPointer[T]] interface {
	types.Controller
	GetDefinition() TypedDefinition[T, P]
	GetTypedIndex(name string) cacheindex.TypedIndex[T]
}

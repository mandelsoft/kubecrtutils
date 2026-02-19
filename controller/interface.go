package controller

import (
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/types"
)

type Controllers = types.Controllers
type Controller = types.Controller

type TypedController[P kubecrtutils.ObjectPointer[T], T any] interface {
	Controller
	GetDefinition() TypedDefinition[P, T]
	GetLocalIndex(name string) cacheindex.TypedIndex[T]
}

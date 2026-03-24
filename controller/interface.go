package controller

import (
	"github.com/mandelsoft/kubecrtutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectMapper[T client.Object, R any] = types.ObjectMapper[T, R]

type ControllerAware[T any] = types.ControllerAware[T]
type ClusterAware[T any] = types.ClusterAware[T]

type Controllers = types.Controllers
type Controller = types.Controller
type ControllerNames = types.ControllerNames

type TypedController[P kubecrtutils.ObjectPointer[T], T any] interface {
	Controller
	GetDefinition() TypedDefinition[P, T]
	GetLocalIndex(name string) cacheindex.TypedIndex[T]
}

type ClusterNames = types.ClusterNames

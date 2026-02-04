package controller

import (
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types"
)

type _controllers struct {
	internal.Group[types.Controller]
}

func NewControllers() types.Controllers {
	return &_controllers{internal.NewGroup[types.Controller]("controller")}
}

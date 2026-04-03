package types

import (
	"github.com/mandelsoft/kubecrtutils/internal"
)

type ComponentNames = NameSet

type ComponentFilter interface {
	GetUsedComponents(ConstraintContext) ComponentNames
}

type Component interface {
	GetName() string
}

type Components interface {
	internal.Group[Component]
	IsDisabled(name string) bool
}

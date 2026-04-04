package types

import (
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types/plain"
)

type ComponentNames = plain.ComponentNames

type ComponentFilter interface {
	GetUsedComponents(ConstraintContext) ComponentNames
}

type Component interface {
	GetName() string
	GetEffective() Component
}

type Components interface {
	internal.Group[Component]
	IsDisabled(name string) bool
}

package types

import (
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types/plain"
	"github.com/mandelsoft/logging"
)

type ComponentNames = plain.ComponentNames

type ComponentFilter interface {
	GetUsedComponents(ConstraintContext) ComponentNames
}

type Component interface {
	logging.Logger

	GetName() string
	GetEffective() Component

	GetComponent(name string) Component
	GetIndex(name string) Index
}

type Components interface {
	internal.Group[Component]
	IsDisabled(name string) bool
}

package component

import (
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/internal"
	"github.com/mandelsoft/kubecrtutils/types"
)

////////////////////////////////////////////////////////////////////////////////

type ComponentNames = types.ComponentNames

type Components = types.Components

type components struct {
	internal.Group[Component]
	disabled set.Set[string]
}

var _ Components = (*components)(nil)

func NewComponents() Components {
	return newComponents()
}

func newComponents() *components {
	return &components{Group: internal.NewGroup[Component]("cluster"), disabled: set.New[string]()}
}

func (c *components) IsDisabled(name string) bool {
	return c.disabled.Has(name)
}

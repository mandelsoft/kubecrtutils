package types

import (
	"github.com/mandelsoft/kubecrtutils/controller/constraints/types"
)

type Activation = types.Activation

type Constraint interface {
	Match(ConstraintContext) (Activation, error)
}

type Constraints interface {
	Constraint
	Len() int
	Add(constraints ...Constraint) Constraints
	Clone() Constraints
}

type ConstraintContext interface {
	Names() ControllerNames
	Selected() ControllerNames
	Has(name ...string) bool
	HasAny(name ...string) bool
	GetGroup(name string) ControllerNames

	WithSelectedSet(names ControllerNames) ConstraintContext
	WithSelected(selected ...string) ConstraintContext
}

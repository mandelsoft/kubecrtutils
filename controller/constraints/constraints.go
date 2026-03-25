package constraints

import (
	"slices"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
)

type ControllerNames = set.Set[string]

type Constraint interface {
	Match(*Context, ControllerNames) error
}

type Constraints interface {
	Constraint
	Len() int
	Add(constraints ...Constraint) Constraints
	Clone() Constraints
}

type _constraints []Constraint

func New() Constraints {
	return &_constraints{}
}

func (c *_constraints) Len() int {
	if c == nil {
		return 0
	}
	return len(*c)
}

func (c *_constraints) Clone() Constraints {
	if c == nil {
		return nil
	}
	return generics.PointerTo(slices.Clone(*c))
}

func (c *_constraints) Add(constraints ...Constraint) Constraints {
	*c = append(*c, constraints...)
	return c
}

func (c *_constraints) Match(ctx *Context, set ControllerNames) error {
	for _, constraint := range *c {
		if err := constraint.Match(ctx, set); err != nil {
			return err
		}
	}
	return nil
}

type Context struct {
	names  ControllerNames
	groups map[string]ControllerNames
}

func NewContext(def activationopts.ControllerSet) *Context {
	ctx := &Context{
		names:  set.New[string](def.GetNames()...),
		groups: make(map[string]ControllerNames),
	}
	groups := def.GetGroups()
	for name := range groups {
		result := ControllerNames{}
		ctx.groups[name] = Closure(name, groups, result, ControllerNames{})
	}
	return ctx
}

func (c *Context) Names() ControllerNames {
	return c.names
}

func (c *Context) GetGroup(name string) ControllerNames {
	return c.groups[name]
}

func Closure(name string, groups map[string][]string, result ControllerNames, handled ControllerNames) ControllerNames {
	if handled.Contains(name) {
		return result
	}
	handled.Add(name)
	grp := groups[name]
	if len(grp) == 0 {
		result.Add(name)
	} else {
		for _, n := range grp {
			Closure(n, groups, result, handled)
		}
	}
	return result
}

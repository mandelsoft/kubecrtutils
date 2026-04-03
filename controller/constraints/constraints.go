package constraints

import (
	"maps"
	"slices"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/set"
	mytypes "github.com/mandelsoft/kubecrtutils/controller/constraints/types"
	"github.com/mandelsoft/kubecrtutils/types"
)

type ControllerNames = types.ControllerNames
type ControllerSet = types.ControllerSet

type Constraint = types.Constraint
type Constraints = types.Constraints

type Activation = mytypes.Activation

const (
	Yes       = mytypes.Yes
	No        = mytypes.No
	NoOpinion = mytypes.NoOpinion
)

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

func (c *_constraints) Match(ctx Context) (Activation, error) {
	return And(*c...).Match(ctx)
}

type Context = types.ConstraintContext

type _context struct {
	all      ControllerNames
	groups   map[string]ControllerNames
	selected ControllerNames
}

func NewContext(def ControllerSet) Context {
	ctx := &_context{
		all:    set.New[string](def.GetNames()...),
		groups: make(map[string]ControllerNames),
	}
	groups := def.GetGroups()
	for name := range groups {
		result := ControllerNames{}
		ctx.groups[name] = Closure(name, groups, result, ControllerNames{})
	}
	return ctx
}

func (c *_context) WithSelected(selected ...string) Context {
	return &_context{
		all:      set.Clone(c.all),
		groups:   maps.Clone(c.groups),
		selected: set.New[string](selected...),
	}
}

func (c *_context) WithSelectedSet(names ControllerNames) Context {
	return &_context{
		all:      set.Clone(c.all),
		groups:   maps.Clone(c.groups),
		selected: set.Clone(names),
	}
}

func (c *_context) Names() ControllerNames {
	return c.all
}

func (c *_context) Selected() ControllerNames {
	return c.selected
}

func (c *_context) Has(name ...string) bool {
	return c.selected.Has(name...)
}

func (c *_context) HasAny(name ...string) bool {
	return c.selected.HasAny(name...)
}

func (c *_context) GetGroup(name string) ControllerNames {
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

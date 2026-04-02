package constraints

import (
	"maps"
	"slices"

	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/set"
)

type ControllerNames = set.Set[string]

type ControllerSet interface {
	GetNames() []string
	GetGroups() map[string][]string
}

type Activation int

const Yes Activation = 1
const No Activation = -1
const NoOpinion Activation = 0

func (a Activation) String() string {
	switch a {
	case No:
		return "No"
	case Yes:
		return "Yes"
	case NoOpinion:
		return "NoOpinion"
	}
	return "Unknown"
}

var orMatrix = [3][3]Activation{
	{No, No, Yes},
	{No, NoOpinion, Yes},
	{Yes, Yes, Yes},
}

func (a Activation) Or(b Activation) Activation {
	return orMatrix[a+1][b+1]
}

var andMatrix = [3][3]Activation{
	{No, No, No},
	{No, NoOpinion, Yes},
	{No, Yes, Yes},
}

func (a Activation) And(b Activation) Activation {
	return andMatrix[a+1][b+1]
}

func (a Activation) Not() Activation {
	return -a
}

type Constraint interface {
	Match(*Context) (Activation, error)
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

func (c *_constraints) Match(ctx *Context) (Activation, error) {
	return And(*c...).Match(ctx)
}

type Context struct {
	all      ControllerNames
	groups   map[string]ControllerNames
	selected ControllerNames
}

func NewContext(def ControllerSet) *Context {
	ctx := &Context{
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

func (c *Context) WithSelected(selected ...string) *Context {
	return &Context{
		all:      set.Clone(c.all),
		groups:   maps.Clone(c.groups),
		selected: set.New[string](selected...),
	}
}

func (c *Context) WithSelectedSet(names ControllerNames) *Context {
	return &Context{
		all:      set.Clone(c.all),
		groups:   maps.Clone(c.groups),
		selected: set.Clone(names),
	}
}

func (c *Context) Names() ControllerNames {
	return c.all
}

func (c *Context) Selected() ControllerNames {
	return c.selected
}

func (c *Context) Has(name ...string) bool {
	return c.selected.Has(name...)
}

func (c *Context) HasAny(name ...string) bool {
	return c.selected.HasAny(name...)
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

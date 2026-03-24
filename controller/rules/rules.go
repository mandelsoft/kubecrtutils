package rules

import (
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	"github.com/mandelsoft/kubecrtutils/types"
)

type Rule interface {
	Match(*Context, types.ControllerNames) error
}

type Rules interface {
	Rule
	Len() int
	Add(rules ...Rule) Rules
}

type _rules []Rule

func New() Rules {
	return &_rules{}
}

func (r *_rules) Len() int {
	if r == nil {
		return 0
	}
	return len(*r)
}

func (r *_rules) Add(rules ...Rule) Rules {
	*r = append(*r, rules...)
	return r
}

func (r *_rules) Match(ctx *Context, set types.ControllerNames) error {
	for _, rule := range *r {
		if err := rule.Match(ctx, set); err != nil {
			return err
		}
	}
	return nil
}

type Context struct {
	names  types.ControllerNames
	groups map[string]types.ControllerNames
}

func NewContext(def activationopts.ControllerSet) *Context {
	ctx := &Context{
		names:  set.New[string](def.GetNames()...),
		groups: make(map[string]types.ControllerNames),
	}
	groups := def.GetGroups()
	for name := range groups {
		result := types.ControllerNames{}
		ctx.groups[name] = Closure(name, groups, result, types.ControllerNames{})
	}
	return ctx
}

func (c *Context) Names() types.ControllerNames {
	return c.names
}

func (c *Context) GetGroup(name string) types.ControllerNames {
	return c.groups[name]
}

func Closure(name string, groups map[string][]string, result types.ControllerNames, handled types.ControllerNames) types.ControllerNames {
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

package rules

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/kubecrtutils/types"
)

type _complete struct {
	groups []string
}

func Complete(grps ...string) Rule {
	return &_complete{groups: grps}
}

func (r *_complete) Match(ctx *Context, cur types.ControllerNames) error {
	for _, group := range r.groups {
		g := ctx.GetGroup(group)
		if g == nil {
			return fmt.Errorf("group %q used but not declared", group)
		}
		if len(g) != len(g.Intersection(cur)) {
			return fmt.Errorf("group %q must be complete [%s]", group, strings.Join(maputils.OrderedKeys(g), ", "))
		}
	}
	return nil
}

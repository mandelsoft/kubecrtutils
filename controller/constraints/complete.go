package constraints

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/maputils"
)

type _complete struct {
	groups []string
}

func Complete(grps ...string) Constraint {
	return &_complete{groups: grps}
}

func (r *_complete) Match(ctx *Context) (Activation, error) {
	for _, group := range r.groups {
		g := ctx.GetGroup(group)
		if g == nil {
			return NoOpinion, fmt.Errorf("group %q used but not declared", group)
		}
		if len(g) != len(g.Intersection(ctx.Selected())) {
			return NoOpinion, fmt.Errorf("group %q must be complete [%s]", group, strings.Join(maputils.OrderedKeys(g), ", "))
		}
	}
	return NoOpinion, nil
}

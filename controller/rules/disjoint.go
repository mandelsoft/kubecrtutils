package rules

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/types"
)

type _disjoint struct {
	groups []string
}

func Disjoint(grps ...string) Rule {
	return &_disjoint{groups: grps}
}

func (r *_disjoint) Match(ctx *Context, cur types.ControllerNames) error {

	for i := 0; i < len(r.groups); i++ {
		found := ""
		for j, group := range r.groups[i:] {
			g := ctx.GetGroup(group)
			if g == nil {
				return fmt.Errorf("group %q used but not declared", group)
			}
			members := g.Intersection(cur)
			if members.Len() > 0 {
				if found != "" {
					return fmt.Errorf("use only controllers either in group %q or %q", found, group)
				}
				found = group
				i = j
			}
		}
	}
	return nil
}

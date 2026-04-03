package constraints

import (
	"fmt"
)

type _disjoint struct {
	groups []string
}

func Disjoint(grps ...string) Constraint {
	return &_disjoint{groups: grps}
}

func (r *_disjoint) Match(ctx Context) (Activation, error) {

	for i := 0; i < len(r.groups); i++ {
		found := ""
		for j, group := range r.groups[i:] {
			g := ctx.GetGroup(group)
			if g == nil {
				return NoOpinion, fmt.Errorf("group %q used but not declared", group)
			}
			members := g.Intersection(ctx.Selected())
			if members.Len() > 0 {
				if found != "" {
					return Yes, fmt.Errorf("use only controllers either in group %q or %q", found, group)
				}
				found = group
				i = j
			}
		}
	}
	return NoOpinion, nil
}

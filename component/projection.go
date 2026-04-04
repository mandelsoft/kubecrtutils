package component

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/mapping"
)

func (cl *components) Map(mapping mapping.Mappings, names ComponentNames) (Components, error) {
	n := NewComponents()

	for local := range names {
		global := mapping.Map(local)
		c := cl.Get(global)
		if c == nil {
			return nil, fmt.Errorf("global component %q for %q not defined", global, local)
		}
		n.Add(NewAlias(local, c))
	}
	return n, nil
}

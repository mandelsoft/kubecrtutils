package cluster

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/mapping"
)

func (cl *_clusters) Map(mapping mapping.Mappings, names ClusterNames) (Clusters, error) {
	n := NewClusters()

	for local := range names {
		global := mapping.Map(local)
		c := cl.Get(global)
		if c == nil {
			return nil, fmt.Errorf("global cluster %q for %q not defined", global, local)
		}
		n.Add(NewAlias(local, c))
	}
	return n, nil
}

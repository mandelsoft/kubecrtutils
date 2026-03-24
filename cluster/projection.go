package cluster

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/types"
)

func Map(clusters Clusters, mapping types.Mappings, names ClusterNames) (Clusters, error) {
	n := NewClusters()

	for local := range names {
		global := mapping.Map(local)
		c := clusters.Get(global)
		if c == nil {
			return nil, fmt.Errorf("global cluster %q for %q not defined", global, local)
		}
		n.Add(NewAlias(local, c))
	}
	return n, nil
}

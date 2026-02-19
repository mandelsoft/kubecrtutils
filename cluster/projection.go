package cluster

import (
	"fmt"

	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Map(clusters Clusters, mapping types.Mappings, names sets.Set[string]) (Clusters, error) {
	n := NewClusters()

	for local := range names {
		global := mapping.Map(local)
		c := clusters.Get(global)
		if c == nil {
			return nil, fmt.Errorf("global cluster %q for %q not defined", global, local)
		}
		n.Add(NewClusterLikeAlias(local, c))
	}
	return n, nil
}

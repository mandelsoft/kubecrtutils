package owner

import (
	"github.com/mandelsoft/kubecrtutils/types"
)

func MatcherFor(c types.ClusterEquivalent) ClusterMatcher {
	return func(clusterId string) string {
		if cl := c.GetClusterById(clusterId); cl != nil {
			return cl.GetName()
		}
		return ""
	}
}

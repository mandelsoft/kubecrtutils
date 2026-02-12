package owner

import (
	"github.com/mandelsoft/kubecrtutils/types"
)

func MatcherFor(c types.ClusterEquivalent) ClusterMatcher {
	return func(clusterId string) (string, bool) {
		ok := clusterId == c.GetId()
		if cl := c.GetClusterById(clusterId); cl != nil {
			return cl.GetName(), ok
		}
		return "", false
	}
}

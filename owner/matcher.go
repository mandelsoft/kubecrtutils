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

func LocalMatcher(localName, localId string) ClusterMatcher {
	return func(id string) (string, bool) {
		if id == "" {
			id = localId
		}
		if id == localId {
			return localName, localId == id
		}
		return "", false
	}
}

func MatcherForClusters(c types.Clusters, localId string) ClusterMatcher {
	return func(id string) (string, bool) {
		if id == "" {
			if localId == "" {
				return "", true // default local cluster
			}
			id = localId
		}
		cl := c.GetClusterById(id)
		if cl != nil {
			return cl.GetName(), id == localId
		}
		return "", false
	}
}

package cluster

import (
	"strings"

	"github.com/mandelsoft/kubecrtutils/cluster/fleet/fpi"
)

func Normalize(clusterName string) string {
	if strings.HasPrefix(clusterName, fpi.SEPARATOR) {
		return strings.TrimPrefix(clusterName, fpi.SEPARATOR)
	}
	return clusterName
}

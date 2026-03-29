package log

import (
	"fmt"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils/log/message"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/mandelsoft/logging"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MessageProvider = message.MessageProvider

type clusterinfo struct {
	types.ClusterEquivalent
	base string
}

func ClusterInfo(c types.ClusterEquivalent, base ...string) MessageProvider {
	return clusterinfo{c, general.Optional(base...)}
}

func (c clusterinfo) NormalizeTo(i *[]interface{}) {
	*i = append(*i, "clusterkind"+c.base, c.GetTypeInfo(), "cluster"+c.base, c.GetName(), "clusterid"+c.base, c.GetId(), "clusterinfo"+c.base, c.GetInfo(), "effcluster"+c.base, c.GetEffective().GetName())
}

func (c clusterinfo) Message() string {
	return ClusterInfoMsg(c.base)
}

func ClusterInfoMsg(base ...string) string {
	b := general.Optional(base...)
	return fmt.Sprintf("{{clusterkind%s}} {{cluster%s}}[{{clusterid%s}}] accessing {{clusterinfo%s}}", b, b, b, b)
}

////////////////////////////////////////////////////////////////////////////////

type logical struct {
	clusterinfo
}

func LogicalClusterInfo(c types.ClusterEquivalent, base ...string) MessageProvider {
	return logical{clusterinfo{c, general.Optional(base...)}}
}

func (c logical) Message() string {
	return LogicalClusterInfoMsg(c.base)
}

func LogicalClusterInfoMsg(base ...string) string {
	b := general.Optional(base...)
	return fmt.Sprintf("{{clusterkind%s}} {{cluster%s}}->{{effcluster%s}}[{{clusterid%s}}]", b, b, b, b)
}

////////////////////////////////////////////////////////////////////////////////

type cluster clusterinfo

func Cluster(c types.ClusterEquivalent, base ...string) MessageProvider {
	return cluster{c, general.Optional(base...)}
}

func (c cluster) NormalizeTo(i *[]interface{}) {
	*i = append(*i, "clusterkind"+c.base, c.GetTypeInfo(), "cluster"+c.base, c.GetName(), "effcluster"+c.base, c.GetEffective().GetName())
}

func (c cluster) Message() string {
	return ClusterMsg(c.base)
}

func ClusterMsg(base ...string) string {
	b := general.Optional(base...)
	return fmt.Sprintf("{{kind%s}} {{cluster%s}}->{{effcluster%s}}", b, b, b)
}

////////////////////////////////////////////////////////////////////////////////

func GroupKind(gk schema.GroupKind) logging.KeyValueProvider {
	return KeyValue("groupkind", gk)
}

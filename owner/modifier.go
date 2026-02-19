package owner

import (
	"context"

	"github.com/mandelsoft/kubecrtutils/cluster/clustercontext"
	"github.com/mandelsoft/kubecrtutils/objutils/objfilter"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type modificationHandler struct {
	cluster types.Cluster
	owner   client.Object
	handler Handler
	filter  objfilter.Interface
}

var _ types.ObjectModifier = (*modificationHandler)(nil)

func AddOwnerModifier(ctx context.Context, owner client.Object, f objfilter.Interface, h Handler) types.ObjectModifier {
	return &modificationHandler{
		cluster: clustercontext.ClusterFor(ctx),
		owner:   owner,
		handler: h,
		filter:  f,
	}
}

func (m *modificationHandler) Modify(cluster types.Cluster, obj client.Object) error {
	if !m.Filter(obj) {
		return nil
	}
	return m.handler.SetOwner(m.cluster, m.owner, cluster, obj)
}

func (c *modificationHandler) Filter(obj client.Object) bool {
	if c.filter == nil {
		return true
	}
	return c.filter.Filter(obj)
}

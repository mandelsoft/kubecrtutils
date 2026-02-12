package builder

import (
	"github.com/mandelsoft/kubecrtutils/types"
	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
)

func For(b *mcbuilder.Builder, cl types.ClusterEquivalent) Builder {
	return &_mcbuilder{cl, b}
}

type _mcbuilder struct {
	cluster types.ClusterEquivalent
	builder *mcbuilder.Builder
}

var _ Builder = (*_mcbuilder)(nil)

func (b *_mcbuilder) Named(name string) Builder {
	b.builder = b.builder.Named(name)
	return b
}

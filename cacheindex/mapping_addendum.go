package cacheindex

import (
	"github.com/mandelsoft/kubecrtutils/mapping"
)

func WithMappings(def Definition) *_mapped {
	m := &_mapped{
		Definition: def,
	}
	m.Mapped = mapping.NewBaseMappings[*_mapped](m)
	m.BaseMappable = m.Mapped
	return m
}

type _mapped struct {
	Definition
	mapping.BaseMappable[*_mapped]
	Mapped *mapping.BaseMappings[*_mapped]
}

var _ Definition = (*_mapped)(nil)

func (m *_mapped) MapCluster(tgt string) *_mapped {
	m.Mapped.MapCluster(m.Definition.GetTarget(), tgt)
	return m.Mapped.GetSelf()
}

func (m *_mapped) MapIndex(tgt string) *_mapped {
	m.Mapped.MapIndex(m.Definition.GetTarget(), tgt)
	return m.Mapped.GetSelf()
}

func (m *_mapped) GetTarget() string {
	return m.ClusterMappings().Map(m.Definition.GetTarget())
}

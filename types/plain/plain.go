package plain

import (
	"github.com/mandelsoft/goutils/set"
)

type NameSet = set.Set[string]

func NewNameSet(names ...string) NameSet {
	return set.New[string](names...)
}

type ClusterNames = NameSet
type ComponentNames = NameSet
type ControllerNames = NameSet
type IndexNames = NameSet

package plain

import (
	"github.com/mandelsoft/goutils/set"
)

type NameSet = set.Set[string]

type ClusterNames = NameSet
type ComponentNames = NameSet
type ControllerNames = NameSet
type IndexNames = NameSet

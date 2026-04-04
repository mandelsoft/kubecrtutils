package ctrlmgmt

import (
	"github.com/mandelsoft/kubecrtutils/mapping"
	"github.com/mandelsoft/kubecrtutils/types"
)

type DefinitionProvider[D any] = types.DefinitionProvider[D]

type ControllerManager = types.ControllerManager
type ControllerDefinition = types.ControllerDefinition

type Mappings = mapping.Mappings
type ClusterNames = types.ClusterNames
type ComponentNames = types.ComponentNames

package ctrlmgmt

import (
	"github.com/mandelsoft/kubecrtutils/types"
)

type DefinitionProvider[D any] = types.DefinitionProvider[D]

type ControllerManager = types.ControllerManager
type MappedControllerDefinition = types.MappedControllerDefinition
type ControllerDefinition = types.ControllerDefinition

type Mappings = types.Mappings
type ClusterNames = types.ClusterNames

{{components}}
## Components

A component is an arbitrary element with dependencies to other components, which may use clusters and  carry indices and is startable by the controller manager, e.g. A web server using access to a Kubernetes cluster.

### Definition
{{component-definitions}}

Components can be defined with the following functions:

- `component.Define(<name>, <cluster>, <factory>)`: Definition of a component created by a factory with the following interface:
  ```go
  {{include}{$(root)/../component/def.go}{factory}}
  ```

  The factory can implement the [`Options`]({{options}}) interface
  to request command line flags.

The factory must provide a `ComponentImplementation` implementing the interface

 ```go
 {{include}{$(root)/../types/components.go}{component implementation}}
 ```
It gets access to a component object providing the declared elements.

#### Modifiers

Modifiers are used to declare nested features of the components.

- `UseCluster(<name>)`: declare the usage of an additional cluster.
- `WithActivationConstraint(<constraints>...)`: Activation constraints defined in package [constraints](../controller/constraints).
- `UseComponent(<name>...)`: require access to a component. Those components are automatically activated, if the controller is active.

Activation constraints are used to automatically activate or deactivate components and to validate possible controller selections.

The component has the following interface:
```go
{{include}{$(root)/../types/components.go}{component}}
```

#### Mapping
{{component-mappings}}

When components are orchestrated in a controller manager, their local names (for clusters, indices and components) can be mapped
to names unique in the scope of the controller manager.
This way, component definitions must not be globally aligned to be orchestratable in a controller manager. Instead, the orchestrator
is responsible to configure a globally consistent mapping of local names.

A definition can be wrapped into a `WithMappings(<definition>)` call.
It provides some mapping methods:

- `MapCluster(<local>, <global>)`: map local cluster names
- `MapIndex(<local>, <global>)`: map local (relative) index names
- `MapComponent(<local>, <global>)`: map required local component names

The mapping of [absolute index names]({{indices}}) involves the cluster and index mapping.

#### Interfaces

A component may implement the `manager.Runnable` interface from the controller runtime to get started by the controller manager.
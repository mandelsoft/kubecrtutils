{{manager}}
## Controller Manager

The controller manager is the heart of the library. It is based on
a multi-cluster runtime controller manager and coordinates all
the elements, like [components]({{components}}), [controllers]({{controllers}}) and [indices]({{indices}}).

Like those elements the element orchestration is defined in a declarative way.  Afterwards, it can be instantiated with some simple calls.

### Definition
{{manager-definitions}}

A controller manager is declared with the following function

- `ctrlmgmt.Define(<name>, <cluster>)`: Definition of a controller controller manager for a given main [cluster]({{clusters}}) 

The complete definition interface is as follows:

```go
 {{include}{$(root)/../ctrlmgmt/def.go}{definition}}
```

The definition is used to orchestrate further nested elements intended to be managed by the declared controller manger.

### Modifiers

The following features can be configured for a controller manager definition:


- `WithOwnerHandler(<provider>`: Set default [owner handling]({{owners}}).
- `WithScheme(<scheme>)`:  The main scheme to use. It should contain all resource types required by any of the orchestrated elements.
- `AddControllerRule(<constraints>...)`: activation constraints orchestrated components and controllers.

So far, the following orchstration alements are supported:

- `AddCluster(<cluster definitioin>...)`: a [cluster]({{cluster-definitions}})
- `AddComponent(<component definition>...)`:  a [component]({{component-definitions}})
- `AddController(<controller definitioin>...) a [cntroller]({{controller-definitions}})
- `AddIndex(<index definition>...)`: an [index]({{index-definitions}})

### Running a controller manager

A controller manager definition implements the [`Options`]({{options}}) interface and can be added to an `OptionSet`, together
with other configuration options.

It is then started like in out example by calling

```go
{{include}{../$(simple)/cmd/main.go}{execute everything}}
 ```

It runs the option lifecycle and runs an instantiation of the controller manage in-between. 

### Mapping

The orchestratable element definitions can be implemented completely independent of each other.
Therefore, if not bundled in a single project, every such element can
decide on its own names. If put together, these name sets might be incompatible or inconsistent. Especially for [indices]({{indices}}) the base names are important, because they are assumed to represent dedicated meanings and are used for deduplication.

The task of the orchestration is then not only to compose the set of elements , but also to provide a consistent name set for the controller manager definition.

This can be achieved by name mappings. Every orchestrated element can be wrapped into an appropriate `<package>.WithMappings` call, which
the allows to configure name mapping for involved elements.
This way, for example, a consistent (and may be simplified) set of logical clusters can be provided on the manager level.
For those cluster, then command line options are provided, automatically.

Mappings are avaiable for
- [*clusters*]({{cluster-mappings}})
- [*indices*]({{index-mappings}})
- [*components*]({{component-mappings}})
- [*controllers*]({{controller-mappings}})



## Configuration by Options
{{managerconfig}}

An option in the used main option set may implement
the manager configuration interface

```go
{{include}{../$(root)/options/manageropts/config.go}{config interface}}
```

When the manager is instantiated it scans for implementations of
this interface and calls it on all found options to apply additional configurations.

Avaiulable options provided by the library:
- `metricsopts`: Options to enable the metrics server for the controller manager (uses `tlsopts`)

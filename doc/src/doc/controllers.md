{{controllers}}
## Controllers

Controllers bundle a reconciler for a main resource, some locally used indices and triggers. Triggers
are watches on other resources (or clusters) triggering reconcilation of the main resource.

### Controller Definitions

Controllers can be defined with the following functions:

- `controller.Define[*<resource>,<resource>](<name>, <cluster>, <factory>)`: Definition of a controller reconciling a *resource* using a reconcier provided by a factory with the following interface:
  ```go
  {{include}{$(root)/../controller/def.go}{reconciler factory}}
  ```

  The factory can implement the [`Options`]({{options}}) interface 
  to request command line flags.

- `controller.DefineByFunc[*<resource>,<resource>](<name>, <cluster>, <function>)`: Definition of a controller reconciling a *resource* using a reconcier provided by a factory function with the following interface:

  ```go
  {{include}{$(root)/../controller/def.go}{reconciler factory function}}
  ```

The *cluster* is the cluster used to watch the given *resource*.
We need both types, the pointer type of the resource (implementing `client.Object` and the struct type) to offer the possibility to declate typed indices of the main resource. An index delivers the slice of struct types.

The complete interface is as follows:

```go
 {{include}{$(root)/../controller/def.go}{controller definition}}
```

#### Modifiers

Modifiers are used to declare nested features of the controllers.

- `InGroup(<name>)`: add controller to a controller group
- `UseCluster(<name>)`: declare the usage of an additional cluster.
- `WithFinalizer(<name>)`: Change the default finalizer name
- `WithPredicates(<preds> ...)`: Predicates
- `WithActivationConstraint(<constraints>...)`: Activation constraints defined in package [constraints](../controller/constraints).
- `UseComponent(<name>...)`: require access to a component. Those components are automatically activated, if the controller is active.

Activation constraints are used to automatically activate or deactivate controller and to validate possible controller selections.

Reconciler factories have access to the actual controller settings.
Here, they can access the declared indices and clusters, as well as options.

The controller has the following interface:
```go
{{include}{$(root)/../types/controllers.go}{controller}}
```

##### Triggers

A trigger is a watch on another resource (and potentially on another cluster) used to rigger the reconsilation of the maon resource on the controller's main cluster.

- `AddTrigger(<trigger definitions>)`: define triggers

There are two predefined trigger definition types:

- `ResourceTrigger[*<resource>](<map function>, <description>)`: trigger reconcilation with an additional resource watch mapping resources to reconcilation requests. The mapping supports a fleet member target.

- `ResourceTriggerByFactory[*<resource>](<map function factory>, <description>)`: Use a controller- and cluster-aware factory. Use `handler.Lift*` variants to reduce the awareness.

- `LocalResourceTrigger[*<resource>](<map function>, <description>)`:
   variant for effective local cluster triggering.

- `ResourceTriggerByFactory[*<resource>](<map function factory>, <description>)`: variant for effective local cluster triggering, only.

The first variants enable cross fleet member triggering. The second variants trigger reconcilation on the cluster/fleet member.

- `OwnerTrigger[*<resource>]()`: Trigger owners of given resource.

Trigger definitions have an optional modifier:

- `OnCluster(<cluster>)`: by default the watch is established on the controllers main cluster ([equivalent]({{clusters}})). With this modifier it can be redirected to another declared logical cluster.
- 
##### Indices

There are two different kind of indices:
- indices on the main resource on the main cluster
- indices on other resource and/or clusters
  - indices defined by the controller
  - indices imported by the controller (expected to be defined somewhere else).

{{include}{indices.md}{index names}}
{{include}{indices.md}{deduplication}}

When accessing indices from the controller. relative names can be used
for indices on the main cluster of the controller.
Other indices can be accessed by using theit global name (still using the local logical cluster names). A name can be composed with `cacheindex.ComposeName`.

If the name of foreign indices are unique in the scope of the controller, in addition to the global names, the relative names can also be used.

- `AddIndex(<relative name>, <typed indexer function>)`: Define an index on the main resource and cluster of the controller.
- `AddIndexByFactory(<relative name>, <typed indexer factory>)`: Define an index on the main resource and cluster of the controller using a factory.

- `AddForeignIndex(<index definitions>)`: use other [index definitions]({{index-definitions}})
- `ImportIndex(<index reference>)` use [index references]({{index-references}}).

A factory (`func(<ctx>, <logger>, <clusters>) (<indexer function>, error)` has access to the actual settings when creating the inder function:
- *options*: use `cacheindex.OptionsFromContext`
- *controller*: use `controller.ControllerFromContext`

#### Mappings

When controllers are orchestrated in a controller manager, their local names (for clusters, indices and components) can be mapped
to names unique in the scope of the controller manager.
This way, controller definition must not be globally aligned to orchestrated in a controller manager. Instead, the orchestrator
is responsible to configure a globally consistent mapping of local names.

A definition can wrapped into a `WithMappings(<definition>)` call.
It provides some mapping methods:

- `MapCluster(<local>, <global>)`: map local cluster names
- `MapIndex(<local>, <global>)`: map local (relative) index names
- `MapComponent(<local>, <global>)`: map required local component names

The mapping of [absolute index names]({{indices}}) involves the cluster and index mapping.

#### Reconciler Support

The uses by default the regular controller runtime for reconcilers.
In case of [fleets]({{clusters}}), the fleet member information is available via the context using 
- `clustercontext.CLusterFor(<ctx>)`: get the cluster object
- `clustercontext.ClusterNameFor(<ctx>)`: get the cluster name

All methods on a `Clusterequivalent` take a context and determine the effective cluster to execute the operation for from the context.

Above this low-level interface, there are three higer level abstractions:

- using a request object bases reconcilation logic (`logic.New`)
- using a request base reconciler using various factories (`factories.New`)
- using low-level request object based reconciler (`reconciler.New`)

All those wrappers bundle the reconcilation environment for 
a dedicated reconcilation request into a `reconciler.ReconcileRequest` object.
It provides access to dynamic information like the effective cluster and the
resource object (as modifiable version and as original version),
but also for shared information like the controller, the reconciler,
the options and explicitly arbitrarily maintained shared information provided by the various factory variants.

The underlying reconciler maps controller-runtime reconcilation requests to such an object. If the resource uses a status field/resource, it automatically checks whether the status has been changed and initiates a status update.

##### `reconciler.New`

This is the basic abstraction. It creates a reconciler factory based on a `reconciler.DefaultRequestFactoryFunc` function, able to create a `reconciler.ReconcileRequest`.

```go
{{include}{$(root)/../controller/controllerutils/reconciler/request.go}{reconcile request}}
```

This request must implement the reconcilation logic

```go
{{include}{$(root)/../controller/controllerutils/reconciler/request.go}{reconcilation logic}}
```

and implement all theother request related method. This is supported
by providing a `reconciler.BaseRequest`, which can be embedded into the final reconcilation request object to implement all those state and methods.

There is a second more general flavor pair `reconciler.NewWithOptions`/`reconciler.DefaultRequestFactoryFuncWithOptions`, which allows to specify an `Options` type, which is added to
the reconciler factory and forwarded to the `baseRequest`-

##### `factories.New`

On top of this abstraction `factories.New` uses three separate factory functions (`Optionfactory`, `SettingsFactory`, and `RequestFactory`),
for options, the shared `Settings` and one for the request creation to
provide an appropriate factory and reconciler.

The provided factory has the interface 

```go
{{include}{$(root)/../controller/controllerutils/reconciler/factories/factory.go}{factory}}
```

and is implemented by a `DefaultFactory` using the factory functions.

A second flavor `factories.NewByFactory` directly take such an implementation

##### `logic.New`

This abstraction is based on the previous one and simplifies all the factory functions.

It only requires to implement the pure logic and shared state by implementing the interface ``

```go
{{include}{$(root)/../controller/controllerutils/reconciler/logic/request.go}{reconcilation logic}}
```

A second flavor `NewWithOptions` addtionally accepts a type parameter for the `Options` type.

Both flavors use a standard implementation for the request, which does not be implemented separately anymore. It forwards
the logic implementation to the shared logic object by passing the request as argument.

This flavor is finally used by ur [walkthrough example]({{walkthrough}})
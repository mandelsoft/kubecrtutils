# A declarative Frontend for the Kubernetes Multi-Cluster Runtime Library

{{variable}{root}{../..}}
{{variable}{simple}{$(root)/examples/simple}}
{{variable}{replicate}{$(simple)/controllers/replicate}}

This project provides a wrapper for the Kubernetes [Multi-CLuster Runtime Library](https://github.com/kubernetes-sigs/multicluster-runtime). It supports a more declarative API for the orchestration of a controller manager.

- *Declarative API*

  This is achieved by strictly distinguishing between code for orchestrating the controller manager and the functional code for the controllers.
  The first part is completely replaced by some declarative-like code and the effective  orchestration is done based on those declarations without further
  use code.

  The second part is provided by factories configured in the declarative part.

  This might look less flexible, but it avoids lots of intransparent boilerplate code for standard use cases. And, if required, implementing an own `Component` can be used to directly access the underlying controller runtime library.

  So far, the following elements are supported:
  - Controllers
  - Indices
  - Arbitrary Components based on the multi-cluster runtime library.


- *Automated Command Line Option Handling*

  Additionally, the handling of [command line options]({{options}}) is also done by the library.
  Factories just need to implement the `Options` API to add command line flags.
  The aggregation to the command line options using a `pflag.FlagSet` is automated by the library.

- *Support for Controllers using multiple clusters*

  The library supports controllers working on multiple clusters by introducing
  logical clusters for functional elements.
  Those elements can be arbitrarily orchestrated into a single controller manager.
  The mapping of logical clusters to physical ones is done based on the
  given command line options and is completely transparent for the elements as long as the library API is used.


- *Transparent Support for Fleet Environments like KCP*

  Controllers can work with fleets similat to single clusters using the
  same API. The library functions implicitly handle the fleet cluster instance
  by using the `context.Context`. 


- *Cross-CLuster/Namespace Owner Handling*

  It is possible to establish, track and get triggered by cross-cluster or cross-namespace ownership relations.

More detailed information is described for the following topics
 - [Options]({{options}})
 - [Clusters]({{clusters}})
 - [Controller Manager]({{manager}})
 - [Controllers]({{controllers}})
 - [Components]({{components}})
 - [Indices]({{indices}})
 - [Owner Handling]({{owners}})

## Walkthrough
{{walkthrough}}

To illustrate the usage and the descriptive power of the library we use a simple
example: we want to create a set of controllers orchestrated into a
controller manager replicating config resources with a particular annotation
into another custer.

There are two directitions:
- *from source to target (called up)*

  here typically the spec part of an object is replicated (in our example the complete object)
- *from target to sour e (called down)*

  here typically the status part is replicated bach to the surce object. For `ConfigMaps` there is no status, but we have to hanle the unexpected deletion.

What we can see already from this specification, there is the need to work with (potentially) different clusters, the source and the target cluster
(which might be identical for special cases).

Therefore, multiple clusters play a central role in this library. 
It uses so-called [*logical clusters*]({{clusters}}). Every element, like a controller may declare and use multiple clusters represented by arbitrary names unique in the context of an element definition.

### Our first controller

So, we start with our first controller, the *replicate* controller replicating an object from source to target.

```go
{{include}{$(simple)/controllers/replicate/def.go}{definition}}
```

We do this by providing a function `Controller`, which creates a `controller.Definition`.
It has some unique name and declares which clusters it should work on, and on which main resource the reconciler should work.

To support typed resource indices at the definition level, the resource must be declared by its pointer and non-pointer type. We want to use the constraint `client.Object` to assure that it is a resource type, which is not implemented by the non-pointer type. But for the index we need the non-pointer type, because the `List` operation requires it for the result type. While it is possible to derive a pointer type from any type, it is not possible in Go to derive the non-pointer type from a pointer type.

Because we want to replicate, we need access to two clusters, the main cluster (for the watch) of the controller is the source cluster (using the constant `SOURCE`). But we need a second one, the target cluster. This is specified by adding `UseCluster`, which declares additional clusters. 

With `AddTrigger(controller.OwnerTrigger[*Resource]().OnCluster(controllers.TARGET))` 
a watch on the target cluster is requested using the [owner information]({{owners}}) on
a target resource to trigger the reconcilation on changes on the replica.

Then we need to describe the reconcilation logic. This is done by providing a factory
able to create a regular `reconcile.Reconciler` from the controller runtime, when the definition is instantiated.

This factory just returns a regular cluster-runtime reconciler featuring the logic. But we want to
use some more comfort and decide to use a support wrapper by using a standard 
factory working on a `support.Request` object. For every reconcilation, such an object is created holding all necessary information required to implement the reconcilation step.

Part of this information is derived from the controller settings and (potential additional options).
Here, for example, we use the effective source and target cluster. These settings are bundled
in a dedicated `Settings` object:

```go
{{include}{$(simple)/controllers/settings.go}{settings}}

```

To control the replication we need some more information, which should be passed
by command line options.
This can be handled completely local to our controller code. We just create
an `Options` type passed as type parameter to our declaration method (`controllers.Options`).

```go
{{include}{$(simple)/controllers/options.go}}
```

We will use a special annotation used to hold a *replication class*. 
Only objects with a dedicated value here, will be replicated.

All this information is packed into the call to `support.NewByLogic` in addition to
our implementation object for the reconcilation logic (`ReconcilationLogic`). The method
then provides a regular factory for the cluster runtime reconciler based
an on object implementing the following interface:

```go
{{include}{$(root)/controller/controllerutils/reconciler/support/reconciler.go}{reconcilation logic}}
```

This object is two-folded:
- It provides the `Settings` from and for the concrete instantiation. This information is
  shared among all reconcilation requests and finally accessible from the used `Request`
  object.

  ```go
  {{include}{$(simple)/controllers/replicate/logic.go}{reconcilation logic}}
  ```

- Implement the reconcilation interface to execute the reconcilation logic for a particular
  request. We will have a look at the implementation [later]({{logic}}).

### The controller manager 

Before we start with our replication logic, we will first create the
main program to run the controller.

We create our main program in [examples/simple/cmd/main.go](../../examples/simple/cmd/main.go).
It consists of three very simple delarative-like parts:

1) Define the scheme you want to use

   ```go
   {{include}{$(simple)/cmd/main.go}{scheme}}
   ```
  
  This is standard coding known from the plain controller runtinme.

2) Configure the elements for the controller manager

   ```go
   {{include}{$(simple)/cmd/main.go}{orchestrate controller manager}}
   ```

  Similar to the controller declaration, it is a declaration refering to our earlier
  declaration methods (like `Controller()`).

3) Configure the options you want to use

   ```
   {{include}{$(simple)/cmd/main.go}{orchestrate general functionality}}
   ```

   Here, some standard options are composed. We want to use the metrics server,
   configure the logging, and we want to be able to activate dedicated controllers.

   All options represent functionality implemented via an `Options` object.  It may offer command line-options, but may do other things, also. It either
   directly executes those features or it implements additional configuration
   interfaces used by other components/options to retrieve configuration. For example,
   it might be able to configure the controller manager (like the metrics option).

   The controller manger definition is basically also such an option, able to instantiate
   a controller manager based on its definitions and potentially other options
   able to further configure it.

4) And finally run the complete configuration

   ```go
   {{include}{$(simple)/cmd/main.go}{execute everything}}
   ```
   
All the flags required by the configured options and all the orchestrated
components will automatically provided to a central `pflag.FlagSet`object.

The command-line flags of our example look as follows:

```
{{execute}{go}{run}{../../examples/simple/cmd/main.go}{--help}}
```

{{logic}}
### The reconcilation logic

We decided not to implement a plain reconciler for the cluster runtime but use
a `Request`-based standard implementation.

The reconcilation logic is implemented by an object implementing the interface 

```go
{{include}{$(root)/controller/controllerutils/reconciler/support/reconciler.go}{reconcilation logic}}
```

As has been seen earlier, it is used to provide a share `Settings` object holding
information shared among all reconcilation requests.

The reconcilatiuon itself is spilt into three methods:
- `Reconcile` for regular resource reconcilation
- `ReconcileDeleting` handle deletion while finalizers are still set
- `ReconcileDeleted` handle the final deletion

All three methods gain access to the reconcilatuon context by a `support.Request`
object. It looks like this:

```go
{{include}{$(root)/controller/controllerutils/reconciler/support/reconciler.go}{request}}
```

which provides access to some shared information found in the `Reconciler` object.

```go
{{include}{$(root)/controller/controllerutils/reconciler/support/reconciler.go}{reconciler}}
```

Here, you find your `Settings` and `Options`.

The dynamic information is found in the embedded `BaseRequest` field. It provides the interface

```go
{{include}{$(root)/controller/controllerutils/reconciler/reconciler.go}{request}}
```

and some important direct fields:

```go
{{include}{$(root)/controller/controllerutils/reconciler/reconciler.go}{request fields}}
```

If there is a status resource, it assumes, that there ia a field `Status`
. If after a reconcilation is done and the status has been changed it is automatically updated using the status resource. This does not need to be done by the reconcilation logic implemented by the controller.

`Finalizer` and `FieldManger` fields describe the values that should/will be used by those purposes.

An [`OwnerHandler`]({{owners}}) can be used to maintain and query owner relationships. This handler
automatically handles cross-namespace and cross-cluster relationships. So, we can use this
to describe the relationship between a replica and its original resource.

Now, we can start implementing our logic.

The reconcilation methods return a `reconcile.Problem` object. It describes
- whether there is a temporary problem, which lead to a rate-limited repetition of the 
  reconcilation
- whether there is a problem with the settings of the resource, which cannot be solved by repeating the reconcilation
- whether there is a reconcilation problem handled somewhere else.

The `reconcile` package provides approprriate constructor methods:

- `Requeue(err error)` requeue in case of an error without reporting it
- `Failed(err error)`  a persistent error
- `TemporaryProblem(err error)` a temporary error solvable by rate-limited reconcilation

The recilation defined in [examples/simple/controllers/repicate/logig.go](examples/simple/controllers/repicate/logig.go)
first checks an annotation set for replicated object, to avoid recursive replicas in case of
 both clusters being identical.

Then it checks the responsibility based on the annotation information provided by the command line options. If not responsible delete a potentiall existing replica.

With

```go
{{include}{$(simple)/controllers/replicate/logic.go}{finalizer}}
```

a finalizer is set on the original object using the FInalizer information from the request object.

With 

```go
{{include}{$(simple)/controllers/replicate/logic.go}{mapping}}
```

the mapping between original and replica names is registered in the shared state.
Hereby, `objutils.GenerateUniqueName` is used to generate a unique name for the 
namespace used to store the replicas.

The [mapping component](examples/simple/controllers/mapping.go) also handle the
namespace vreation and deletion, by keeping track of existent replications 
for this namespace.

It is important to know, that every cluster has an assigned abstract identity.
By default, this is the name of the cluster as defined by the controller manager.
But to use clusters independent of controllers and controller managers,
command line options are provided to configure an identity for clusters.
This identity is used for the owner reference, but here also for 
the generation of the target namespace name. Whenever some persistent
names depend on a cluster, always the cluster identity should be used,
instead of the (logical) cluster name local to the controller environment.

With

```go
{{include}{$(simple)/controllers/replicate/logic.go}{prepare}}
```

the replica state is prepared. An owner information is established
with `err := r.Reconciler.SetOwner(r.Cluster, r.Object, s.Target, newp)`

With

```go
{{include}{$(simple)/controllers/replicate/logic.go}{prepare}}
```

it would be possible to transfer ther status from an existing target object,
not yet used for our demo resource.


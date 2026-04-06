# A declarative Frontend for the Kubernetes Multi-Cluster Runtime Library

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

  Additionally, the handling of command line options is also done by the library.
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



## The Main Function

Every element intended to be used in a controller manager it declared
similar to this example:

```go
const Index = "myindex"

func Controller() controller.Definition {
	return controller.Define[*corednsv1alpha1.HostedZone](common.ControllerHostedzone, "dataplane", &ReconcilerFactory{}).
		UseCluster("runtime").
		InGroup("functional").
		AddIndex(Index, Indexer).
		ImportIndex(cacheindex.Ref[*corednsv1alpha1.CoreDNSEntry, corednsv1alpha1.CoreDNSEntry](IndexKeyEntryZone, "dataplane")).
		AddTrigger(
			controller.OwnerTrigger[*appsv1.Deployment]().OnCluster("runtime"),
			controller.OwnerTrigger[*corev1.Secret]().OnCluster("runtime"),
			controller.LocalResourceTriggerByFactory[*corev1.Secret](secretTriggerFactory).OnCluster("runtime"),
		)
}
```



## Options

## Clusters

## Indices

## Components

## ross-CLuster/Namespace Owner Handling


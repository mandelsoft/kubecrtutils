{{owners}}
## Owner Handling

When working with multiple clusters a typical use case is to 
establish cross cluster relationships, which should be registerd
at some resource. The most common relationship is the ownership, expressing that a resource object has been generated to implement
another resource.

Kubernetes itself supports only namespace-local owner references,
even cross-namespace references are not available.

The package [`owner`](../../owner) provides an abstraction
for handing arbitrary cross-cluster and cross-namespace reference in addition to the well-known local references.

The interface (package `types`)

```go
{{include}{../$(root)/types/owners.go}{owner handler}}
```

is used for such handlers (in package `owner` the name is just `owner.Handler`). It can be used to establish and retrieve relations.


The `ClusterMatcher` is used to map [cluster identities]({{cluster-identities}}) to cluster names used in this library.
It also determines, whether a particular identity should be considered for an opertaion on the handler.

The [`ClusterEquivalent`]({{clusters}})] interface also supports this interface.

### Standard Handler

There is a default implementation falling back to Kubernetes owner references  for namespace-local relations and using an annotation
for cross-references.

It can be configured with an annotation type.
The default used for owner references is

```go
{{include}{../$(root)/owner/anno.go}{default annotation}}
```

The annotation type 

```go
{{include}{../$(root)/owner/anno.go}{annotation type}}
```

handles the representation of references as annotation.
A default type can be created for any annotation name.

The annotations stores reference information using the group id, kind, namespace and name.
If required the cluster identity is added.

### Handling of cross-cluster references
{{cluster-identities}}

All those solutions have to persist cluster identities. This library
uses logical clusters names, on the level of the controller manager elements and the controller manager itself. Those names are
defined by the local coding. Therefore, they are not applicable, if multiple controllers should work together on a common set of clusters.

This is the reason, why the cluster abstraction in this library features
separate cluster identities for every used cluster. It is defaulted 
by the cluster name, but can be externally managed by command line options. This way, the deployer of controllers can provide a consistent view on cluster identities across a set of controller (managers).
{{clusters}}
## Clusters

Clusters are the central element of the library. It manages the mapping of logic clusters
used by the elements orchestrated into a controller manager to physical clusters configured
via the command line options.

Elements like [indices]({{indices}}), [components]({{components}}) or [controllers]({{controllers}})
are declared for clusters identified by same logical cluster name.
During the instantiation the various factories get access to a mapping of declared names to real cluster objects.

When aggregating those elements into a [controller manager]({{manager}}) the cluster names are also aggregated and mapped to logical clusters declared for the controller manager.

During the startup phase of the controller manager those names are mapped to physical cluster objects generated according to the command line settings.

### Cluster Abstraction

One goal of the library is to support regular clusters as well as cluster fleets, like [KCP](https://kcp.io) in a mostly transparent manner.

To achieve this the type `ClusterEquivalent` is used as common interface for both incarnations
of a cluster environment. The logical cluster names always relate to cluster equivalents.
It offers a possibility to determone the type and the interface for the concrete variant:
- `CLuster`: is the interface for regular Kubernetes clusters
- `Fleet`: is the interface for sets of clusters (like KCO workspaces)

Both incarnations can be transparently handled by most of the elements and functions.
Controllers and indices are defined for `ClusterEquivalent`s instad of regular Kubernetes clusters.

Reconcilers either use the regular cluster runtime interface or a mor sofisticated request interface. The actually involved cluster is provided by `context.Context`. Also
generic operations lifted to the `ClusterEquivalient`like `Get` or `List` implicitly evaluate the context to determine the effective fleet instance to work on.

Therefore, typical controllers don't need to knwo whetjer theay are working on a `Cluster` or `Fleet` variant of a `ClusterEquivalent`.

```go
{{include}{$(root)/../types/clusters.go}{cluster equivalent}}
```

The `ClusterEquivalent` is the common interface also implemented by the variants `Cluster` and `Fleet`.  It can be used fortypical operations on clusters.

#### Cluster Names and Identities

Every cluster, logical and physical ones, has name. For fleets the names are extended by 
a sub name identifying the aggregated cluster in the fleet.
The name extension is never changed when mapping a cluster name, only the name of the outer `ClusterEquivalent`. `ClusterEquivalent`s backed by a regular cluster have the same name than the underlying cluster.

In addition to the name a cluster also features an identity. The identity is used to
persist clusters identities outside of the library, for example, as owner references 
in Kubernetes data planes. The identity is introduced to enable the usage of globally unique identities independent of the controllers or controller managers used.

By default, the identity is set to the name of the underlying physical cluster. But there are dedicated options to configure dedicated identities when starting the controller manager.


#### Regular Clusters

Regular clusters implement the interface `cluster.Cluster`:

```go
{{include}{$(root)/../types/clusters.go}{cluster}}
```

#### Fleets

Regular clusters implement the interface `cluster.Fleet`:
```go
{{include}{$(root)/../types/clusters.go}{fleet}}
```

A fleet provides access to the clusters actually aggregated in the fleet.
So far, only a fleet implementation for *KCP* is supported.


### Configuring Access

For every cluster declared for the controller manager a set of
command line options is generated:

- `kubeconfig` *string*:   path to standard kubeconfig
- `kubeconfig-context` *string* (optional): context used together with kubeconfig
- `kubeconfig-identity` *string* (optional): identity used together with kubeconfig

For a KCP fleet an additional option is required:

- `kubeconfig-endpointslice` *string*   endpointslice used together with kubeconfig for APIExport of a KCP fleet

This is the standard option set used for the defaulr cluster.
Additionally, the same set of options is generated for every configured cluster by prefixing the option with the cluster name. In our example, these are `source-` and `target-`.

If the definition enables the usage f a fleet, the required fleet options are added.

The complete configuration discovery for handled by a simple rules option engine (package [`rules`](../cluster/config)).
It enables composing rule sets for intended clusters, using a selection of possible rules:

- context specification
- environment settings
- in-cluster config
- identity configuration
- kubeconfig options (with special names, like `in-cluster`)
- special rules for fleet configuration.

By evaluating the rules against a set of command line flags
provide an appropriate configuration object, including the required rest config.

This is done by implementing the [`Options` interface]({{options}}) by the `Rules` object aggregating the rule set for a dedicated cluster.

### Definition
{{cluster-definitions}}

Logical clusters (cluster equivalents) can be defined using two definition flavors:

- `cluster.Define(<name>, <description>)`: Definition of a regular logical Kubernetes cluster

- `cluster.DefineFleet(<name>, <description>, <type>)`: Definition of a logical cluster optionally enabling the fleet feature for the given type. Supported types:
  - `kcp.Type()`: KCP fleet type
  The fleet type enables additional fleet specific options.
  Future development, will change this by introducing a fleet type option, if multiple fleet types should be possible.

#### Modifiers
{{cluster-modifiers}}

A cluster definition allows some modifiers:
- `.WithFallback(<name>)`: If the logical cluster is not configured by the command line, a fallback cluster is used instead. If no fallback is configured, the cluster is required to be configured.

#### Mapping
{{cluster-mappings}}

For cluster definitions there is no explicit `.WithMappings` call.
Nevertheless, there is some other kind of mapping feature.

When defining clusters at the level of the [controller manager]({{manager}}) cluster definitions can basically be mapped to each other by defining [fallbacks]({{cluster-modifiers}}).
All definitions are used, but when evaluating the command line options, undefined clusters can be mapped to other ones, according to their fallback definitions.

#### Remarks

Additionally, a default cluster with the name `cluster.DEFAULT` is always configured. There should be a fallback to this cluster. It is used if no other 
cluster config option is given. It also includes the `in-cluster` rule for configuration. For controllers running in a cluster this cluster is by default configured for accessing the local runtime cluster.
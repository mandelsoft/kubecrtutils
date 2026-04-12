{{indices}}
## Indices

<!--- begin index names --->
Indices use relative and absolute names. A relative name is the name as used in the declarations. The absolute name is composed by the local name and the name of the cluster it is defined for. 
<!--- end index names --->

An absolute name has the form `<relative>:<cluster>`.

### Deduplications

<!--- begin deduplication --->
Indices are deduplicated, considering the base name, resource and cluster. If an index is defined on multiple clusters, that are all mapped to the same physical cluster, the index is created only once.

In Go it is not possible to compare indexer functions, therefore, the base name of an index is meant to uniquely define the meaning independent of the used cluster. When orchestration index defining elements in the controller manager those *meanings* can be aligned with [mappings](indices.md#mapping)
<!--- end deduplication --->

For example, the relative index name describes an indexed feature of a resource. This is specific to the index and resource, but does not depend on the cluster the index is created on.

To bring together local meanings and global meanings required for deduplication, relative names can be [mapped]({{index-mappings}}) when orchestrated into a controller manager.

### Definition
{{index-definitions}}

Indices can be defined with the following functions:

- `cacheindex.Define[*<resource>,<resource>](<name>, <cluster>, <index function>)`: Definition of an index on a cluster with a typed
  indexer function
  ```go
  {{include}{$(root)/../types/indices.go}{indexer func}}
  ```
  
- `cacheindex.DefineByFactory[*<resource>,<resource>](<name>, <cluster>, <index factory>)`: Definition of an index on a cluster with the following factory function. It accepts a context and the mapped clusters.

  The context also provides access to the options and the defining element with appropriate context access functions.
  
{{index-references}}

- `Ref[*<resource>,<resource>](<name>,<cluster>)`: Definition of an
  index reference. Those definitions can be used with `ImportIndex` modifiers of other elements.

### Modifiers

This element has no modifiers.

### Mapping
{{index-mappings}}

Indices are either diretly orchestrated into a controller manager 
or as nested elements of [components]({{components}}) or [controllers]({{controllers}}).
In any case, their local names (for clusters, indices and components) can be mapped
to names unique in the scope of the controller manager.
This way, index definitions must not be globally aligned to be orchestratable in a controller manager. Instead, the orchestrator
is responsible to configure a globally consistent mapping of local names.

A definition can be wrapped into a `WithMappings(<definition>)` call.
It provides some mapping methods:

- `MapCluster(<global>)`: map local index cluster name
- `MapIndex(<global>)`: map local (relative) index name

The mapping of [absolute index names]({{indices}}) involves the cluster and index mapping.

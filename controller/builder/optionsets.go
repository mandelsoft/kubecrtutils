package builder

import (
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/kubecrtutils/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
)

// project represents other forms that we can use to
// send/receive a given resource (metadata-only, unstructured, etc).
type objectProjection int

const (
	// projectAsNormal doesn't change the object from the form given.
	projectAsNormal objectProjection = iota
	// projectAsMetadata turns this into a metadata-only watch.
	projectAsMetadata
)

type ClusterFilterFunc = mcbuilder.ClusterFilterFunc

type untypedWatchesInput interface {
	setPredicates([]predicate.Predicate)
	setObjectProjection(objectProjection)
	setClusterFilter(ClusterFilterFunc)
}

type commonOptions struct {
	predicates       []predicate.Predicate
	objectProjection objectProjection
	clusterFilter    ClusterFilterFunc
}

func (w *commonOptions) setPredicates(predicates []predicate.Predicate) {
	w.predicates = predicates
}

func (w *commonOptions) setObjectProjection(objectProjection objectProjection) {
	w.objectProjection = objectProjection
}

func (w *commonOptions) setClusterFilter(clusterFilter ClusterFilterFunc) {
	w.clusterFilter = clusterFilter
}

func (o *commonOptions) applyFleetFilter(fleet types.Fleet) {
	if o.clusterFilter == nil {
		o.clusterFilter = fleet.Filter
	} else {
		exp := o.clusterFilter
		o.clusterFilter = func(clusterName string, cluster cluster.Cluster) bool {
			return fleet.Filter(clusterName, cluster) && exp(clusterName, cluster)

		}
	}
}

func (o *commonOptions) mapToMCRT() []any {
	var opts []any
	if len(o.predicates) > 0 {
		opts = append(opts, mcbuilder.WithPredicates(o.predicates...))
	}
	if o.objectProjection != 0 {
		opts = append(opts, mcbuilder.OnlyMetadata)
	}
	if o.clusterFilter != nil {
		opts = append(opts, mcbuilder.WithClusterFilter(o.clusterFilter))
	}
	return opts
}

func (o *commonOptions) mapToCRT() []any {
	var opts []any
	if len(o.predicates) > 0 {
		opts = append(opts, builder.WithPredicates(o.predicates...))
	}
	if o.objectProjection != 0 {
		opts = append(opts, builder.OnlyMetadata)
	}
	return opts
}

// forOptions represents the options set of the For method.
type forOptions struct {
	commonOptions
}

func (o *forOptions) mapToMCRT() []mcbuilder.ForOption {
	opts := sliceutils.Convert[mcbuilder.ForOption](o.commonOptions.mapToMCRT())
	return opts
}

func (o *forOptions) mapToCRT() []builder.ForOption {
	opts := sliceutils.Convert[builder.ForOption](o.commonOptions.mapToCRT())
	return opts
}

////////////////////////////////////////////////////////////////////////////////

// ownsOptions represents the option set for the Owns method.
type ownsOptions struct {
	commonOptions
	matchEveryOwner bool
}

func (o *ownsOptions) mapToMCRT() []mcbuilder.OwnsOption {
	opts := sliceutils.Convert[mcbuilder.OwnsOption](o.commonOptions.mapToMCRT())

	if o.matchEveryOwner {
		opts = append(opts, mcbuilder.MatchEveryOwner)
	}
	return opts
}

func (o *ownsOptions) mapToCRT() []builder.OwnsOption {
	opts := sliceutils.Convert[builder.OwnsOption](o.commonOptions.mapToCRT())

	if o.matchEveryOwner {
		opts = append(opts, builder.MatchEveryOwner)
	}
	return opts
}

////////////////////////////////////////////////////////////////////////////////

// watchesOptions represents the information set by Watches method.
type watchesOptions struct {
	commonOptions
}

func (o *watchesOptions) mapToMCRT() []mcbuilder.WatchesOption {
	opts := sliceutils.Convert[mcbuilder.WatchesOption](o.commonOptions.mapToMCRT())
	return opts
}

func (o *watchesOptions) mapToCRT() []builder.WatchesOption {
	opts := sliceutils.Convert[builder.WatchesOption](o.commonOptions.mapToCRT())
	return opts
}

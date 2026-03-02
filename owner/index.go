package owner

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/goutils/transformer"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimtypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Filter[T client.Object] func(obj T, owner Owner) bool

func ConvertFilter[T client.Object](f Filter[T]) Filter[client.Object] {
	return transformer.OptimizedTransform(f, _convertFilter)
}

func _convertFilter[T client.Object](f Filter[T]) Filter[client.Object] {
	return func(obj client.Object, owner Owner) bool {
		return f(any(obj).(T), owner)
	}
}

func Convert[I, O any](in I, conv transformer.Transformer[I, O]) O {
	if c, ok := any(in).(O); ok {
		return c
	} else {
		return conv(in)
	}
}
func (o Filter[T]) applyTo(options *indexOptions) {
	options.filters = append(options.filters, ConvertFilter(o))
}

type ForGroupKind schema.GroupKind

func (o ForGroupKind) applyTo(options *indexOptions) {
	options.gk = generics.PointerTo(schema.GroupKind(o))
}

type indexOptions struct {
	filters []Filter[client.Object]
	gk      *schema.GroupKind
}

type IndexOption interface {
	applyTo(*indexOptions)
}

func Indexer[T client.Object](handler Handler, cmatch ClusterMatcher, opts ...IndexOption) types.IndexerFunc[T] {
	var options indexOptions

	for _, o := range opts {
		o.applyTo(&options)
	}

	mapper := func(o Owner) string {
		return o.AsKey(options.gk == nil)
	}

	if options.gk != nil {
		options.filters = append(options.filters, func(obj client.Object, owner Owner) bool {
			return owner.GroupKind == *options.gk
		})
	}
	if len(options.filters) == 0 {
		return func(obj T) []string {
			return sliceutils.Transform(handler.GetOwners(cmatch, "", obj), mapper)
		}
	}
	return func(obj T) []string {
		owners := handler.GetOwners(cmatch, "", obj)
		return sliceutils.Transform(
			sliceutils.Filter(owners, func(o Owner) bool {
				for _, filter := range options.filters {
					if !filter(obj, o) {
						return false
					}
				}
				return true
			}),
			mapper,
		)
	}
}

////////////////////////////////////////////////////////////////////////////////

func WithCuster(target, current types.Cluster) func(objk client.ObjectKey, gk ...schema.GroupKind) string {
	return func(objk client.ObjectKey, gk ...schema.GroupKind) string {
		key := MapToIndexKey(objk, gk...)
		if target.GetEffective().GetName() != current.GetEffective().GetName() {
			key = target.GetEffective().GetName() + key
		}
		return key
	}
}

func MapToIndexKey(objkey client.ObjectKey, gk ...schema.GroupKind) string {
	key := string(apimtypes.Separator) + objkey.String()
	if len(gk) > 0 {
		key += string(apimtypes.Separator) + gk[0].String()
	}
	return key
}

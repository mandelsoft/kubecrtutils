package cacheindex

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/owner"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ownerOption interface {
	applyToIndex(*options)
}

type options struct {
	gk      *schema.GroupKind
	handler ClustersAware[owner.Handler]
	filter  ClustersAware[owner.Filter[client.Object]]
}

type GroupKind schema.GroupKind

func (o GroupKind) applyToIndex(opts *options) {
	opts.gk = generics.PointerTo(schema.GroupKind(o))
}

type handler struct {
	handler ClustersAware[owner.Handler]
}

func OwnerHandler(h owner.Handler) ownerOption {
	return handler{Lift(h)}
}

func (o handler) applyToIndex(opts *options) {
	opts.handler = o.handler
}

type filter struct {
	filter ClustersAware[owner.Filter[client.Object]]
}

func OwnerFilter[T client.Object](f owner.Filter[T]) ownerOption {
	return filter{Lift(owner.ConvertFilter(f))}
}

func (o filter) applyToIndex(opts *options) {
	opts.filter = o.filter

}

/*
func OwnerIndex[P kubecrtutils.ObjectPointer[T], T any](name string, target string, opts ...ownerOption) Definition {
	var options options

	for _, o := range opts {
		o.applyToIndex(&options)
	}

	idxfunc := owner.MapOwners()
	return &_definition[P, T]{
		Element: internal.NewElement(name),
		target:  target,
		idxfunc: idxfunc,
		proto:   generics.ObjectFor[P](),
	}
}


*/

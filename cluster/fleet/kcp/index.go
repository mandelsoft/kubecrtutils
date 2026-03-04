package kcp

import (
	"context"

	"github.com/kcp-dev/logicalcluster/v3"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (f *Fleet) ListIndexedGlobalKeys(ctx context.Context, obj runtime.Object, index string, key string, opts ...client.ListOption) ([]types.GlobalKey, error) {
	// unfortunately we cannot use unstructurued lists, here, because the cache is typically only configured for
	// structured types, when the controller is working with structured types.
	list, err := objutils.CreateObjectList(obj, f.GetScheme())
	if err != nil {
		return nil, err
	}

	var results []types.GlobalKey
	for _, pp := range f.wrapper.Providers {
		c := pp.GetCache()
		err := c.List(ctx, list, sliceutils.CopyAppend[client.ListOption](opts, client.MatchingFields{index: "*/" + key})...)
		if err != nil {
			return nil, err
		}
		for _, e := range objutils.Items(list) {
			r := types.NewGlobalKey(f.Compose(string(logicalcluster.From(e.(client.Object)))), client.ObjectKeyFromObject(e.(client.Object)))
			results = append(results, r)
		}
	}

	return results, nil
}

func (f *Fleet) ListIndexedGlobalKeysByObjectKey(ctx context.Context, obj runtime.Object, index string, key types.TypedGlobalKey, opts ...client.ListOption) ([]types.GlobalKey, error) {
	list, err := objutils.CreateObjectList(obj, f.GetScheme())
	if err != nil {
		return nil, err
	}

	var results []types.GlobalKey
	for _, pp := range f.wrapper.Providers {
		c := pp.GetCache()
		err := c.List(ctx, list, sliceutils.CopyAppend[client.ListOption](opts, client.MatchingFields{index: "*/" + key.AsKey(true)})...)
		if err == nil {
			for _, e := range objutils.Items(list) {
				r := types.NewGlobalKey(f.Compose(string(logicalcluster.From(e))), client.ObjectKeyFromObject(e))
				results = append(results, r)
			}
		} else {
			return nil, err
		}

		meta.SetList(list, []runtime.Object{})
		err = c.List(ctx, list, sliceutils.CopyAppend[client.ListOption](opts, client.MatchingFields{index: "*/" + key.AsLocalKey().AsKey(true)})...)
		if err == nil {
			for _, e := range objutils.Items(list) {
				c := f.Compose(string(logicalcluster.From(e)))
				if c != key.ClusterName {
					continue
				}
				r := types.NewGlobalKey(c, client.ObjectKeyFromObject(e))
				results = append(results, r)
			}
		} else {
			return nil, err
		}
	}

	return results, nil
}

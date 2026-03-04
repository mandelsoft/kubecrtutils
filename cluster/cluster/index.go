package cluster

import (
	"context"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *_cluster) ListIndexedGlobalKeys(ctx context.Context, obj runtime.Object, index string, key string, opts ...client.ListOption) ([]types.GlobalKey, error) {
	list, err := objutils.CreateObjectList(obj, c.GetScheme())
	if err != nil {
		return nil, err
	}

	var results []types.GlobalKey
	err = c.List(ctx, list, sliceutils.CopyAppend[client.ListOption](opts, client.MatchingFields{index: key})...)
	if err == nil {
		results = make([]types.GlobalKey, objutils.ObjectListLen(list))
		for i, e := range objutils.Items(list) {
			r := types.NewGlobalKey(c.GetName(), client.ObjectKeyFromObject(e))
			results[i] = r
		}
	}
	return results, nil
}

func (c *_cluster) ListIndexedGlobalKeysByObjectKey(ctx context.Context, obj runtime.Object, index string, key types.TypedGlobalKey, opts ...client.ListOption) ([]types.GlobalKey, error) {
	list, err := objutils.CreateObjectList(obj, c.GetScheme())
	if err != nil {
		return nil, err
	}

	var results []types.GlobalKey
	err = c.List(ctx, list, sliceutils.CopyAppend[client.ListOption](opts, client.MatchingFields{index: key.AsKey(true)})...)
	if err == nil {
		results = make([]types.GlobalKey, objutils.ObjectListLen(list))
		for i, e := range objutils.Items(list) {
			r := types.NewGlobalKey(c.GetName(), client.ObjectKeyFromObject(e))
			results[i] = r
		}
	} else {
		return nil, err
	}

	if key.ClusterName == c.GetName() {
		meta.SetList(list, []runtime.Object{})
		err = c.List(ctx, list, sliceutils.CopyAppend[client.ListOption](opts, client.MatchingFields{index: key.AsLocalKey().AsKey(true)})...)
		if err == nil {
			for _, e := range objutils.Items(list) {
				r := types.NewGlobalKey(c.GetName(), client.ObjectKeyFromObject(e))
				results = append(results, r)
			}
		} else {
			return nil, err
		}
	}

	return results, nil
}

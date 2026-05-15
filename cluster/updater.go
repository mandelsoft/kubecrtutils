package cluster

import (
	"fmt"
	"sync"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils/merge"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlser "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	apimachtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/managedfields"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

type Updaters struct {
	lock     sync.Mutex
	manager  string
	updaters map[string]*Updater
}

func NewUpdaters(manager string) *Updaters {
	return &Updaters{updaters: make(map[string]*Updater), manager: manager}
}

func (u *Updaters) Get(name string) *Updater {
	u.lock.Lock()
	defer u.lock.Unlock()
	if u.updaters == nil {
		u.updaters = make(map[string]*Updater)
	}
	e := u.updaters[name]
	if e == nil {
		e = NewUpdater(u.manager)
		u.updaters[name] = e
	}
	return e
}

// Updater describes a manifest template, which can be instantiated
// values under the contraint that all used value combinations
// just set values, but do not generate different object structures.
// When used it builds  a cache for common elements based
// on this structure used to control the patch process.
type Updater struct {
	gvk       *schema.GroupVersionKind
	merger    *merge.ObjectMerger
	converter managedfields.TypeConverter
	defaulted *fieldpath.Set
}

func NewUpdater(manager string) *Updater {
	return &Updater{}
}

func (t *Updater) Update(c Cluster, ctx OperationContext, manifest []byte, mod ...*ModificationInfo) (*unstructured.Unstructured, error) {
	manager := ctx.GetFieldManager()

	if t.merger == nil {
		m, err := merge.NewObjectMerger(c.GetTypeConverter(), c.GetScheme(), manager)
		if err != nil {
			return nil, err
		}
		t.merger = m
	}

	desired := unstructured.Unstructured{}
	dec := yamlser.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode(manifest, nil, &desired)
	if err != nil {
		return nil, err
	}

	// give context the chance to modify object
	err = ctx.Modify(c, &desired)
	if err != nil {
		return nil, fmt.Errorf("failed to modify manifest: %w", err)
	}

	if t.gvk == nil {
		gvk, err := apiutil.GVKForObject(&desired, t.merger.GetScheme())
		if err != nil {
			return nil, err
		}
		t.gvk = &gvk
	}
	if t.converter == nil {
		t.converter, err = t.merger.GetConverter(*t.gvk)
		if err != nil {
			return nil, err
		}
	}

	var liveObj unstructured.Unstructured
	liveObj.SetGroupVersionKind(*t.gvk)
	err = c.Get(ctx, client.ObjectKey{
		// Try to get the current object from the cluster
		Namespace: desired.GetNamespace(),
		Name:      desired.GetName(),
	}, &liveObj)

	if errors.IsNotFound(err) {
		general.Optional(mod...).SetCreated()
		ctx.Info("creating resource {{groupkind}} {{namespace}}/{{name}} in {{cluster}}", "cluster", c.GetName(), "name", desired.GetName(), "namespace", desired.GetNamespace(), "groupkind", desired.GroupVersionKind())
		liveObj = *desired.DeepCopy()
		err = c.Create(ctx, &liveObj, &client.CreateOptions{
			// PATH A: Create if not found
			FieldManager: manager,
		})
		if err != nil {
			return nil, err
		}
		return t.complete(c, ctx, &desired, &liveObj, nil)
	} else if err != nil {
		return nil, err
	}

	patchData, err := t.merger.ComputeSSAPatch(&liveObj, &desired)
	if err != nil {
		return nil, err
	}

	if string(patchData) == "{}" || len(patchData) == 0 {
		ctx.Info("resource {{groupkind}} {{namespace}}/{{name}} in {{cluster}} uptodate", "cluster", c.GetName(), "name", desired.GetName(), "namespace", desired.GetNamespace(), "groupkind", desired.GroupVersionKind())
		return &liveObj, nil // No changes, exit early
	}
	// live, _ := yaml.Marshal(&current)
	// fmt.Printf("*** live:\n%s\n", string(live))
	// fmt.Printf("*** intended:\n%s\n", string(manifest))
	// fmt.Printf("*** patch:\n%s\n", string(patchData))
	rawPatch := client.RawPatch(apimachtypes.MergePatchType, patchData)
	general.Optional(mod...).SetUpdated()
	ctx.Info("apply patch for resource {{groupkind}} {{namespace}}/{{name}} in {{cluster}}", "cluster", c.GetName(), "name", desired.GetName(), "namespace", desired.GetNamespace(), "groupkind", desired.GroupVersionKind(), "patch", string(patchData))
	err = c.Patch(ctx, &liveObj, rawPatch, &client.PatchOptions{
		FieldManager: manager,
	})
	if err != nil {
		return nil, err
	}
	return t.complete(c, ctx, &desired, &liveObj, nil)
}

func (t *Updater) complete(c Cluster, ctx OperationContext, desired, liveObj *unstructured.Unstructured, desiredFields *fieldpath.Set) (*unstructured.Unstructured, error) {
	var err error

	if desiredFields == nil {
		_, desiredFields, err = t.merger.GetInfo(t.converter, desired)
		if err != nil {
			return nil, err
		}
	}
	if liveObj == nil {
		var live unstructured.Unstructured
		live.SetGroupVersionKind(desired.GroupVersionKind())
		// don't use the cache to get the effective object omitting the cache round trip
		err = c.GetAPIReader().Get(ctx, client.ObjectKey{
			// Try to get the current object from the cluster
			Namespace: desired.GetNamespace(),
			Name:      desired.GetName(),
		}, &live)
		liveObj = &live
	}

	ownedFields, err := t.merger.GetOwnedFields(liveObj)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("desired fields:\n %s\n", desiredFields.String())
	// fmt.Printf("effectively owned fields:\n %s\n", ownedFields.String())
	defaulted := ownedFields.Difference(desiredFields)
	if t.merger.GetDefaulted() != nil {
		defaulted = defaulted.Union(t.merger.GetDefaulted())
	}
	// fmt.Printf("defaulted fields:\n %s\n", defaulted.String())
	ctx.Info("remember defaulted fields", "defaulted", defaulted.String())
	t.merger = t.merger.WithDefaulted(defaulted)
	return liveObj, nil
}

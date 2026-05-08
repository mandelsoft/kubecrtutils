package merge

import (
	"bytes"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

func (m *ObjectMerger) ComputeSSAPatch(
	live runtime.Object,
	desired runtime.Object,
) ([]byte, error) {

	gvk, err := apiutil.GVKForObject(live, m.scheme)
	if err != nil {
		return nil, err
	}

	typeConverter, err := m.converters.GetConverter(gvk)
	if err != nil {
		return nil, err
	}

	// 1. Convert live and desired into typed values
	//    (typed = schema-aware, knows list merge keys etc.)
	liveTyped, err := typeConverter.ObjectToTyped(live)
	if err != nil {
		return nil, fmt.Errorf("converting live to typed: %w", err)
	}

	desiredTyped, err := typeConverter.ObjectToTyped(desired)
	if err != nil {
		return nil, fmt.Errorf("converting desired to typed: %w", err)
	}

	// 2. Extract the fieldset this manager currently owns from live
	//    (from .metadata.managedFields)
	liveObj, ok := live.(metav1.Object)
	if !ok {
		return nil, fmt.Errorf("live object does not implement metav1.Object")
	}

	ownedFields := fieldpath.NewSet()
	for _, entry := range liveObj.GetManagedFields() {
		if entry.Manager != m.managerName {
			continue
		}
		fs := &fieldpath.Set{}
		if err := fs.FromJSON(bytes.NewReader(entry.FieldsV1.Raw)); err != nil {
			return nil, fmt.Errorf("parsing managed fields: %w", err)
		}
		ownedFields = ownedFields.Union(fs)
	}

	// Compute the new fieldset from desired
	desiredFields, err := desiredTyped.ToFieldSet()
	if err != nil {
		return nil, fmt.Errorf("extracting desired fieldset: %w", err)
	}

	// Merge desired onto live, scoped to owned fields:
	// - for fields in desiredTyped: take value from desired
	result, err := liveTyped.Merge(desiredTyped)
	if err != nil {
		return nil, fmt.Errorf("merging typed: %w", err)
	}

	// Also handle removed fields:
	// fields previously owned but no longer in desired → remove from result
	removedFields := ownedFields.Difference(desiredFields)
	if !removedFields.Empty() {
		result = result.RemoveItems(removedFields)
		if err != nil {
			return nil, fmt.Errorf("removing dropped fields: %w", err)
		}
	}

	// Check if anything actually changed
	comparison, err := liveTyped.Compare(result)
	if err != nil {
		return nil, fmt.Errorf("comparing typed objects: %w", err)
	}

	fmt.Printf("compare: %s\n", comparison.String())

	if comparison.IsSame() {
		return nil, nil // no patch needed
	}

	// 5. Serialize live and merged result, produce JSON merge patch
	liveJSON, err := json.Marshal(live)
	if err != nil {
		return nil, err
	}

	resultObj, err := typeConverter.TypedToObject(result)
	if err != nil {
		return nil, fmt.Errorf("converting result to object: %w", err)
	}

	resultJSON, err := json.Marshal(resultObj)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreateMergePatch(liveJSON, resultJSON)
	if err != nil {
		return nil, err
	}

	return patch, nil
}

package cacheindex_test

import (
	"github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/kubecrtutils/cacheindex"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type TestType struct {
	v1.TypeMeta `json:",inline"`
	v1.ObjectMeta
	Spec Spec `json:"spec"`
}

type Spec struct {
	ParentRef string `json:"parentRef"`
}

func (t *TestType) GetObjectKind() schema.ObjectKind {
	// TODO implement me
	panic("implement me")
}

func (t *TestType) DeepCopyObject() runtime.Object {
	// TODO implement me
	panic("implement me")
}

var _ = ginkgo.Describe("Field Indexer Test Environment", func() {
	ginkgo.It("field", func() {
		idx := testutils.Must(cacheindex.FieldIndexer[*TestType]("obj.spec.parentRef"))

		var obj TestType

		obj.Spec.ParentRef = "reference"
		Expect(idx(&obj)).To(Equal([]string{"reference"}))
	})
})

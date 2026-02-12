package owner_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/kubecrtutils/owner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Owner Annotation", func() {

	var Id = IdA
	var OwnerGroup = "core"
	var OwnerKind = "secret"
	var Namespace = "namespace"
	var Name = "name"

	typ := owner.DefaultAnnotationType()

	Context("cross namespace", func() {
		var a owner.Annotation

		BeforeEach(func() {
			a = typ.CrossNamespaceAnnotation(OwnerGroup, OwnerKind, Namespace, Name)
		})

		It("basic", func() {
			Expect(a.String()).To(Equal(OwnerGroup + owner.DEFAULT_SEPARATOR + OwnerKind + owner.DEFAULT_SEPARATOR + Namespace + owner.DEFAULT_SEPARATOR + Name))
			Expect(a.Name()).To(Equal(Name))
			Expect(a.Namespace()).To(Equal(Namespace))
			Expect(a.Kind()).To(Equal(OwnerKind))
			Expect(a.Group()).To(Equal(OwnerGroup))
			Expect(a.ClusterId(Id)).To(Equal(Id))
		})

		It("ForCluster", func() {
			Expect(a.ForCluster(Id).String()).To(Equal(Id + owner.DEFAULT_SEPARATOR + OwnerGroup + owner.DEFAULT_SEPARATOR + OwnerKind + owner.DEFAULT_SEPARATOR + Namespace + owner.DEFAULT_SEPARATOR + Name))
			Expect(a.String()).To(Equal(OwnerGroup + owner.DEFAULT_SEPARATOR + OwnerKind + owner.DEFAULT_SEPARATOR + Namespace + owner.DEFAULT_SEPARATOR + Name))
		})

		It("Match", func() {
			Expect(a.Match(IdA, owner.MatcherFor(cluster1), schema.GroupKind{Group: OwnerGroup, Kind: OwnerKind})).To(Equal("A"))
			Expect(a.Match(IdA, owner.MatcherFor(cluster1), schema.GroupKind{Group: OwnerGroup, Kind: "other"})).To(Equal(""))
			Expect(a.Match(IdB, owner.MatcherFor(cluster1), schema.GroupKind{Group: OwnerGroup, Kind: OwnerKind})).To(Equal(""))
		})

		It("parses", func() {
			p := Must(typ.Get(a.Put(nil)))
			Expect(p.Name()).To(Equal(Name))
			Expect(p.Namespace()).To(Equal(Namespace))
			Expect(p.Kind()).To(Equal(OwnerKind))
			Expect(p.Group()).To(Equal(OwnerGroup))
			Expect(p.ClusterId()).To(Equal(""))

			Expect(p.String()).To(Equal(a.String()))
		})
	})

	Context("cross cluster", func() {
		var a owner.Annotation

		BeforeEach(func() {
			a = typ.CrossClusterAnnotation(Id, OwnerGroup, OwnerKind, Namespace, Name)
		})

		It("basic", func() {
			Expect(a.String()).To(Equal(Id + owner.DEFAULT_SEPARATOR + OwnerGroup + owner.DEFAULT_SEPARATOR + OwnerKind + owner.DEFAULT_SEPARATOR + Namespace + owner.DEFAULT_SEPARATOR + Name))
			Expect(a.Name()).To(Equal(Name))
			Expect(a.Namespace()).To(Equal(Namespace))
			Expect(a.Kind()).To(Equal(OwnerKind))
			Expect(a.Group()).To(Equal(OwnerGroup))
			Expect(a.ClusterId("wrong")).To(Equal(Id))
		})

		It("ForCluster", func() {
			Expect(a.ForCluster("other").String()).To(Equal("other" + owner.DEFAULT_SEPARATOR + OwnerGroup + owner.DEFAULT_SEPARATOR + OwnerKind + owner.DEFAULT_SEPARATOR + Namespace + owner.DEFAULT_SEPARATOR + Name))
			Expect(a.String()).To(Equal(Id + owner.DEFAULT_SEPARATOR + OwnerGroup + owner.DEFAULT_SEPARATOR + OwnerKind + owner.DEFAULT_SEPARATOR + Namespace + owner.DEFAULT_SEPARATOR + Name))
		})

		It("Match", func() {
			Expect(a.Match(IdB, owner.MatcherFor(cluster1), schema.GroupKind{Group: OwnerGroup, Kind: OwnerKind})).To(Equal("A"))
			Expect(a.Match(IdB, owner.MatcherFor(cluster2), schema.GroupKind{Group: OwnerGroup, Kind: OwnerKind})).To(Equal(""))
			Expect(a.Match(IdB, owner.MatcherFor(cluster1), schema.GroupKind{Group: OwnerGroup, Kind: "other"})).To(Equal(""))
		})

		It("parses", func() {
			p := Must(typ.Get(a.Put(nil)))
			Expect(p.Name()).To(Equal(Name))
			Expect(p.Namespace()).To(Equal(Namespace))
			Expect(p.Kind()).To(Equal(OwnerKind))
			Expect(p.Group()).To(Equal(OwnerGroup))
			Expect(p.ClusterId()).To(Equal(IdA))

			Expect(p.String()).To(Equal(a.String()))
		})
	})

	Context("hierarchical name", func() {
		var a owner.Annotation

		BeforeEach(func() {
			a = typ.CrossClusterAnnotation("nested/id", OwnerGroup, OwnerKind, Namespace, Name)
		})

		It("basic", func() {
			Expect(a.String()).To(Equal("nested/id/core/secret/namespace/name"))
			Expect(a.Name()).To(Equal(Name))
			Expect(a.Namespace()).To(Equal(Namespace))
			Expect(a.Kind()).To(Equal(OwnerKind))
			Expect(a.Group()).To(Equal(OwnerGroup))
			Expect(a.ClusterId()).To(Equal("nested/id"))
		})

		It("parses", func() {
			p := Must(typ.Get(a.Put(nil)))
			Expect(p.Name()).To(Equal(Name))
			Expect(p.Namespace()).To(Equal(Namespace))
			Expect(p.Kind()).To(Equal(OwnerKind))
			Expect(p.Group()).To(Equal(OwnerGroup))
			Expect(p.ClusterId()).To(Equal("nested/id"))

			Expect(p.String()).To(Equal(a.String()))
		})
	})
})

package owner_test

import (
	"github.com/mandelsoft/goutils/generics"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/owner"
	"github.com/mandelsoft/kubecrtutils/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type ref struct {
	clusterName string
	obj         *sigclient.ObjectKey
}

func Key(obj sigclient.Object) *sigclient.ObjectKey {
	if obj == nil {
		return nil
	}
	return generics.PointerTo(sigclient.ObjectKeyFromObject(obj))
}

func RefO(cluster types.Cluster, obj sigclient.Object) *ref {
	k := Key(obj)
	if k == nil {
		return nil
	}
	return &ref{clusterName: cluster.GetName(), obj: k}
}

func Ref(clusterName string, obj *sigclient.ObjectKey) *ref {
	if obj == nil {
		return nil
	}
	return &ref{clusterName, obj}
}

func CheckAnno(obj sigclient.Object, val string) {
	if val == "" {
		ExpectWithOffset(1, obj.GetAnnotations()).To(BeNil())
	} else {
		ExpectWithOffset(1, obj.GetAnnotations()).NotTo(BeNil())
		v := obj.GetAnnotations()[owner.DEFAULT_ANNOTATION_NAME]
		ExpectWithOffset(1, v).To(Equal(val))
	}
}

var IdA = "IdA"
var IdB = "IdB"
var cluster1 = NewCluster("A", IdA)
var cluster2 = NewCluster("B", IdB)
var clusterN = NewCluster("nested/A", "IdNested/A")

var _ = Describe("Owner Test Environment", func() {
	var _owner sigclient.Object
	var _slaveDefault sigclient.Object
	var _slaveOther sigclient.Object
	var gvkOwner schema.GroupVersionKind
	var gvkSlave schema.GroupVersionKind

	BeforeEach(func() {
		_owner = &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "owner",
			},
		}

		_slaveDefault = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "secret",
			},
		}

		_slaveOther = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "namespace",
				Name:      "secret",
			},
		}

		gvkOwner, _ = apiutil.GVKForObject(_owner, clientgoscheme.Scheme)
		gvkSlave, _ = apiutil.GVKForObject(_slaveDefault, clientgoscheme.Scheme)
	})

	handler := owner.NewHandlerWithScheme(clientgoscheme.Scheme)
	match1 := owner.MatcherFor(cluster1)
	match2 := owner.MatcherFor(cluster2)

	Context("local", func() {

		It("same namespace", func() {
			MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveDefault))
			CheckAnno(_slaveDefault, "")

			Expect(Ref(handler.GetOwner(match1, cluster1, _slaveDefault, gvkOwner.GroupKind()))).To(Equal(RefO(cluster1, _owner)))

			Expect(Ref(handler.GetOwner(match1, cluster1, _slaveDefault, gvkSlave.GroupKind()))).To(BeNil())
			Expect(Ref(handler.GetOwner(match2, cluster1, _slaveDefault, gvkOwner.GroupKind()))).To(BeNil())

			Expect(handler.GetOwners(match1, cluster1.GetId(), _slaveDefault)).To(Equal(
				[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
			))
		})

		It("cross namespace", func() {
			MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveOther))
			CheckAnno(_slaveOther, "core/Service/default/owner")

			Expect(Ref(handler.GetOwner(match1, cluster1, _slaveOther, gvkOwner.GroupKind()))).To(Equal(RefO(cluster1, _owner)))

			Expect(Ref(handler.GetOwner(match1, cluster1, _slaveOther, gvkSlave.GroupKind()))).To(BeNil())
			Expect(Ref(handler.GetOwner(match2, cluster1, _slaveOther, gvkOwner.GroupKind()))).To(BeNil())

			Expect(handler.GetOwners(match1, cluster1.GetId(), _slaveOther)).To(Equal(
				[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
			))
		})
	})

	Context("remote", func() {

		It("same namespace", func() {
			MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveDefault))
			CheckAnno(_slaveDefault, "IdA/core/Service/default/owner")

			Expect(Ref(handler.GetOwner(match1, cluster2, _slaveDefault, gvkOwner.GroupKind()))).To(Equal(RefO(cluster1, _owner)))
			Expect(Ref(handler.GetOwner(match1, cluster2, _slaveDefault, gvkSlave.GroupKind()))).To(BeNil())

			Expect(handler.GetOwners(match1, cluster1.GetId(), _slaveDefault)).To(Equal(
				[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
			))
		})

		It("cross namespace", func() {
			MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveOther))
			CheckAnno(_slaveOther, "IdA/core/Service/default/owner")

			Expect(Ref(handler.GetOwner(match1, cluster2, _slaveOther, gvkOwner.GroupKind()))).To(Equal(RefO(cluster1, _owner)))
			Expect(Ref(handler.GetOwner(match1, cluster2, _slaveOther, gvkSlave.GroupKind()))).To(BeNil())

			Expect(handler.GetOwners(match1, cluster1.GetId(), _slaveOther)).To(Equal(
				[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
			))
		})
	})

	Context("local matcher", func() {
		Context("local", func() {
			It("same namespace", func() {
				MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveDefault))

				Expect(handler.GetOwners(owner.LocalMatcher("", ""), "", _slaveDefault)).To(Equal(
					[]owner.Owner{{"", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
				))

				Expect(handler.GetOwners(owner.LocalMatcher("A", IdA), "", _slaveDefault)).To(Equal(
					[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
				))
			})

			It("cross namespace", func() {
				MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveOther))

				Expect(handler.GetOwners(owner.LocalMatcher("", ""), "", _slaveOther)).To(Equal(
					[]owner.Owner{{"", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
				))

				Expect(handler.GetOwners(owner.LocalMatcher("A", IdA), "", _slaveOther)).To(Equal(
					[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
				))
			})
		})

		Context("remote", func() {

			It("same namespace", func() {
				MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveDefault))

				Expect(handler.GetOwners(owner.LocalMatcher("", ""), "", _slaveDefault)).To(BeNil())
				Expect(handler.GetOwners(owner.LocalMatcher("B", "IdB"), "IdB", _slaveDefault)).To(BeNil())

				Expect(handler.GetOwners(owner.LocalMatcher("A", IdA), "IdB", _slaveDefault)).To(Equal(
					[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
				))
			})

			It("cross namespace", func() {
				MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveOther))

				Expect(handler.GetOwners(owner.LocalMatcher("", ""), "", _slaveOther)).To(BeNil())
				Expect(handler.GetOwners(owner.LocalMatcher("B", "IdB"), "IdB", _slaveOther)).To(BeNil())

				Expect(handler.GetOwners(owner.LocalMatcher("A", IdA), "IdB", _slaveOther)).To(Equal(
					[]owner.Owner{{"A", sigclient.ObjectKey{Namespace: "default", Name: "owner"}, schema.GroupKind{Group: "core", Kind: "Service"}}},
				))
			})
		})
	})

	Context("indexer", func() {
		clusters := cluster.NewClusters()
		clusters.Add(cluster1)
		clusters.Add(cluster2)

		It("for gk (match)", func() {
			indexfunc := owner.Indexer[sigclient.Object](handler, owner.MatcherForClusters(clusters, ""), owner.ForGroupKind(schema.GroupKind{"core", "Service"}))

			MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveDefault))

			Expect(indexfunc(_slaveDefault)).To(Equal(
				[]string{"/default/owner"},
			))
		})

		It("for gk (no match)", func() {
			indexfunc := owner.Indexer[sigclient.Object](handler, owner.MatcherForClusters(clusters, ""), owner.ForGroupKind(schema.GroupKind{"core", "ConfigMap"}))

			MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveDefault))

			Expect(indexfunc(_slaveDefault)).To(BeNil())
		})

		Context("non-cluster-aware", func() {
			indexfunc := owner.Indexer[sigclient.Object](handler, owner.MatcherForClusters(clusters, ""))

			Context("local", func() {
				It("same namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveDefault))

					Expect(indexfunc(_slaveDefault)).To(Equal(
						[]string{"/default/owner/Service.core"},
					))
				})

				It("cross namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster1, _slaveOther))

					Expect(indexfunc(_slaveOther)).To(Equal(
						[]string{"/default/owner/Service.core"},
					))
				})
			})

			Context("remote", func() {
				It("same namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveDefault))

					Expect(indexfunc(_slaveDefault)).To(Equal(
						[]string{"A/default/owner/Service.core"},
					))
				})

				It("cross namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveOther))

					Expect(indexfunc(_slaveOther)).To(Equal(
						[]string{"A/default/owner/Service.core"},
					))
				})
			})
		})

		Context("cluster-aware", func() {
			indexfunc := owner.Indexer[sigclient.Object](handler, owner.MatcherForClusters(clusters, "IdB"))

			Context("local", func() {
				It("same namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster2, _owner, cluster2, _slaveDefault))

					Expect(indexfunc(_slaveDefault)).To(Equal(
						[]string{"B/default/owner/Service.core"},
					))
				})

				It("cross namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster2, _owner, cluster2, _slaveOther))

					Expect(indexfunc(_slaveOther)).To(Equal(
						[]string{"B/default/owner/Service.core"},
					))
				})
			})

			Context("remote", func() {
				It("same namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveDefault))

					Expect(indexfunc(_slaveDefault)).To(Equal(
						[]string{"A/default/owner/Service.core"},
					))
				})

				It("cross namespace", func() {
					MustBeSuccessful(handler.SetOwner(cluster1, _owner, cluster2, _slaveOther))

					Expect(indexfunc(_slaveOther)).To(Equal(
						[]string{"A/default/owner/Service.core"},
					))
				})
			})
		})
	})
})

package owner

import (
	"github.com/mandelsoft/goutils/funcs"
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ClusterMatcher = types.ClusterMatcher
type Handler = types.OwnerHandler
type Owner = types.Owner

var NewOwner = types.NewOwner

type standard struct {
	annoType AnnotationType
	scheme   *runtime.Scheme
}

type HandlerProvider interface {
	GetOwnerHandler(scheme types.SchemeProvider) Handler
}

type _provider struct {
}

var DefaultProvider = &_provider{}

func (p *_provider) GetOwnerHandler(scheme types.SchemeProvider) Handler {
	return NewHandler(scheme)
}

func NewHandler(scheme types.SchemeProvider, annos ...AnnotationType) Handler {
	return NewHandlerWithScheme(scheme.GetScheme(), annos...)
}

func NewHandlerWithScheme(scheme *runtime.Scheme, annos ...AnnotationType) Handler {
	t := general.Optional(annos...)
	if t == nil {
		t = StandardAnnotationType
	}
	return &standard{annoType: t, scheme: scheme}
}

func (h *standard) SetOwner(cluster types.Cluster, owner client.Object, target types.Cluster, slave client.Object) error {
	sameCluster := cluster.IsSameAs(target)
	// check for local ref
	if sameCluster && owner.GetNamespace() == slave.GetNamespace() {
		return controllerutil.SetControllerReference(owner, slave, h.scheme)
	}
	gvk, err := apiutil.GVKForObject(owner, h.scheme)

	if err != nil {
		return err
	}
	if gvk.Group == "" {
		gvk.Group = "core"
	}
	o := h.annoType.CrossNamespaceAnnotation(gvk.Group, gvk.Kind, owner.GetNamespace(), owner.GetName())
	if cluster != nil && !cluster.IsSameAs(target) {
		o = o.ForCluster(cluster.GetId())
	}
	objutils.ModifyAnnotations(slave, o.Put)
	return nil
}

func (h *standard) GetOwner(cmatch ClusterMatcher, target types.Cluster, obj client.Object, kind schema.GroupKind) (string, *client.ObjectKey) {
	eq := funcs.Second(cmatch(target.GetId()))
	if eq {
		for _, r := range obj.GetOwnerReferences() {
			if r.Kind != kind.Kind {
				continue
			}
			gv, _ := schema.ParseGroupVersion(r.APIVersion)
			if gv.Group == kind.Group || (kind.Group == "core" && gv.Group == "") {
				return target.GetName(), &client.ObjectKey{Name: r.Name, Namespace: obj.GetNamespace()}
			}
		}
	}

	if kind.Group == "" {
		kind.Group = "core"
	}

	// [cluster /] group / kind / namespace / name
	a, err := h.annoType.Get(obj.GetAnnotations())
	if a == nil || err != nil {
		return "", nil
	}
	cname := a.Match(target.GetId(), cmatch, kind)
	if cname == "" {
		return "", nil
	}
	return cname, generics.PointerTo(a.ObjectKey())
}

func (h *standard) GetOwners(cmatch ClusterMatcher, targetId string, obj client.Object) []Owner {
	var result []Owner
	n, eq := cmatch(targetId)
	// index functions do not have access to the actual cluster, therefore we use the empty name here to
	// denote the actual (target) cluster
	if n != "" || (n == "" && eq) {
		for _, r := range obj.GetOwnerReferences() {
			gv, _ := schema.ParseGroupVersion(r.APIVersion)
			g := gv.Group
			if g == "" {
				g = "core"
			}
			result = append(result, NewOwner(
				n,
				client.ObjectKey{Name: r.Name, Namespace: obj.GetNamespace()},
				schema.GroupKind{Group: g, Kind: r.Kind},
			))
		}
	}

	// [cluster /] group / kind / namespace / name
	a, err := h.annoType.Get(obj.GetAnnotations())
	if a == nil || err != nil {
		return result
	}

	n, eq = cmatch(a.ClusterId(targetId))
	if n != "" || (eq && n == "") {
		result = append(result, NewOwner(
			n,
			a.ObjectKey(),
			a.GroupKind(),
		))
	}
	return result
}

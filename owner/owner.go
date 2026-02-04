package owner

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/kubecrtutils/objutils"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const DEFAULT_REMOTE_NAME = "cross-cluster.io/owner-id"

type ClusterMatcher func(clusterId string) (clusterName string)

type Handler interface {
	SetOwner(cluster types.Cluster, owner client.Object, target types.Cluster, slave client.Object) error
	GetOwner(cmatch ClusterMatcher, target types.Cluster, obj client.Object, kind schema.GroupKind) (string, *client.ObjectKey)
}

type standard struct {
	property string
	scheme   *runtime.Scheme
}

func NewHandler(scheme *runtime.Scheme, property ...string) Handler {
	return &standard{property: general.OptionalNonZeroDefaulted(DEFAULT_REMOTE_NAME, property...), scheme: scheme}
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
	o := fmt.Sprintf("%s/%s/%s/%s", gvk.Group, gvk.Kind, owner.GetNamespace(), owner.GetName())
	if cluster != nil && !cluster.IsSameAs(target) {
		o += "/" + cluster.GetId()
	}
	objutils.SetAnnotation(slave, h.property, o)
	return nil
}

func (h *standard) GetOwner(cmatch ClusterMatcher, target types.Cluster, obj client.Object, kind schema.GroupKind) (string, *client.ObjectKey) {
	cid := cmatch(target.GetId())
	if cid != "" {
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

	// group / kind / namespace / name / cluster
	a := objutils.GetAnnotation(obj, h.property)
	if a == "" {
		return "", nil
	}
	fields := strings.Split(a, "/")
	cname := h.match(cid, cmatch, fields, kind)
	if cname == "" {
		return "", nil
	}
	return cname, &client.ObjectKey{Namespace: fields[2], Name: fields[3]}
}

func (l *standard) match(localId string, matcher ClusterMatcher, fields []string, kind schema.GroupKind) (clusterName string) {
	if fields[0] == kind.Group && fields[1] == kind.Kind {
		switch len(fields) {
		case 4:
			return localId
		case 5:
			return matcher(fields[4])
		}
	}
	return ""
}

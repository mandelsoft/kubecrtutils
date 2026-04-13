package types

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	apimtypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NormalizeKeyString(fleet Fleet, localClusterName string, key string) string {
	if strings.HasPrefix(key, string(apimtypes.Separator)) {
		return fleet.Compose(localClusterName) + key
	}
	return key
}

type GlobalKey struct {
	ClusterName string
	apimtypes.NamespacedName
}

func NewGlobalKey(clusterName string, obj apimtypes.NamespacedName) GlobalKey {
	return GlobalKey{
		ClusterName:    clusterName,
		NamespacedName: obj,
	}
}

func ParseGlobalKey(key string) (*GlobalKey, error) {
	fields := strings.Split(key, string(apimtypes.Separator))
	if len(fields) != 3 {
		return nil, fmt.Errorf("invalid key: %s", key)
	}
	return &GlobalKey{
		ClusterName: fields[0],
		NamespacedName: apimtypes.NamespacedName{
			Namespace: fields[1],
			Name:      fields[2],
		},
	}, nil
}

func (o GlobalKey) GetClusterName() string {
	return o.ClusterName
}

func (o GlobalKey) GetObjectKey() client.ObjectKey {
	return o.NamespacedName
}

func (o GlobalKey) AsLocalKey() GlobalKey {
	o.ClusterName = ""
	return o
}

func (o GlobalKey) AsKey() string {
	key := o.ClusterName + string(apimtypes.Separator)
	return key + o.NamespacedName.String()
}

func (o GlobalKey) String() string {
	return o.AsKey()
}

type TypedGlobalKey struct {
	GlobalKey
	schema.GroupKind
}

func ParseTypedGlobalKey(key string) (*TypedGlobalKey, error) {
	fields := strings.Split(key, string(apimtypes.Separator))
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid key: %s", key)
	}
	return &TypedGlobalKey{
		GlobalKey: GlobalKey{
			ClusterName: fields[0],
			NamespacedName: apimtypes.NamespacedName{
				Namespace: fields[3],
				Name:      fields[4],
			},
		},
		GroupKind: schema.GroupKind{
			Group: fields[1],
			Kind:  fields[2],
		},
	}, nil
}

func NewGlobalTypedKey(clusterName string, obj apimtypes.NamespacedName, gk schema.GroupKind) TypedGlobalKey {
	return TypedGlobalKey{
		GlobalKey: GlobalKey{
			ClusterName:    clusterName,
			NamespacedName: obj,
		},
		GroupKind: gk,
	}
}

func (o TypedGlobalKey) GetGroupKind() schema.GroupKind {
	return o.GroupKind
}

func (o TypedGlobalKey) AsLocalKey() TypedGlobalKey {
	o.ClusterName = ""
	return o
}

func (o TypedGlobalKey) AsKey(useGK bool) string {
	key := o.ClusterName + string(apimtypes.Separator)
	if useGK && o.Kind != "" && o.Group != "" {
		key += o.GroupKind.String() + string(apimtypes.Separator)
	}
	return key + o.NamespacedName.String()
}

func (o TypedGlobalKey) String() string {
	return o.AsKey(true)
}

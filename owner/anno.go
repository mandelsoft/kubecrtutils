package owner

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/funcs"
	"github.com/mandelsoft/goutils/general"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimtypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DEFAULT_SEPARATOR = string(apimtypes.Separator)

// --- begin default annotation ---

const DEFAULT_ANNOTATION_NAME = "cross-cluster.io/owner-id"

// --- end default annotation ---

// --- begin annotation type ---

// AnnotationType handles anootations used to handle the persistence
// of object relations in annotations.
type AnnotationType interface {
	Get(map[string]string) (Annotation, error)

	CrossNamespaceAnnotation(group, kind, namespace, name string) Annotation
	CrossClusterAnnotation(clusterid, group, kind, namespace, name string) Annotation
}

// --- end annotation type ---

var StandardAnnotationType = DefaultAnnotationType()

type Annotation interface {
	Put(map[string]string) map[string]string

	ForCluster(id string) Annotation

	String() string
	ClusterId(def ...string) string
	Group() string
	Kind() string
	Namespace() string
	Name() string

	GroupKind() schema.GroupKind
	ObjectKey() client.ObjectKey

	Match(localId string, matcher ClusterMatcher, kind schema.GroupKind) (clusterName string)
}

type _AnnotationType struct {
	separator string
	name      string
}

func DefaultAnnotationType(name ...string) AnnotationType {
	return &_AnnotationType{name: general.OptionalNonZeroDefaulted(DEFAULT_ANNOTATION_NAME, name...), separator: DEFAULT_SEPARATOR}
}

func (t *_AnnotationType) Get(annos map[string]string) (Annotation, error) {
	if len(annos) == 0 {
		return nil, nil
	}
	v, ok := annos[t.name]
	if !ok {
		return nil, nil
	}
	a := strings.Split(v, t.separator)
	if len(a) < 4 {
		return nil, fmt.Errorf("invalid anotation format for owner annotation %q", t.name)
	}

	return _Annotation{name: t.name, fields: a}, nil
}

func (t *_AnnotationType) CrossNamespaceAnnotation(group, kind, namespace, name string) Annotation {
	return _Annotation{name: t.name, fields: []string{group, kind, namespace, name}}
}

func (t *_AnnotationType) CrossClusterAnnotation(clusterId, group, kind, namespace, name string) Annotation {
	return _Annotation{name: t.name, fields: []string{clusterId, group, kind, namespace, name}}
}

type _Annotation struct {
	name   string
	fields []string
}

////////////////////////////////////////////////////////////////////////////////

func (a _Annotation) Put(m map[string]string) map[string]string {
	if m == nil {
		m = make(map[string]string)
	}
	m[a.name] = a.String()
	return m
}

func (a _Annotation) ForCluster(id string) Annotation {
	return _Annotation{name: a.name, fields: append([]string{id}, a.fields[len(a.fields)-4:]...)}
}

func (a _Annotation) String() string {
	return strings.Join(a.fields, DEFAULT_SEPARATOR)
}

func (a _Annotation) ClusterId(def ...string) string {
	if len(a.fields) == 4 {
		return general.Optional(def...)
	}
	return strings.Join(a.fields[0:len(a.fields)-4], DEFAULT_SEPARATOR)
}

func (a _Annotation) Group() string {
	return a.fields[len(a.fields)-4]
}

func (a _Annotation) Kind() string {
	return a.fields[len(a.fields)-3]
}

func (a _Annotation) Namespace() string {
	return a.fields[len(a.fields)-2]
}

func (a _Annotation) Name() string {
	return a.fields[len(a.fields)-1]
}

func (a _Annotation) GroupKind() schema.GroupKind {
	return schema.GroupKind{Group: a.Group(), Kind: a.Kind()}
}

func (a _Annotation) ObjectKey() client.ObjectKey {
	return client.ObjectKey{Name: a.Name(), Namespace: a.Namespace()}
}

func (a _Annotation) Match(localId string, matcher ClusterMatcher, kind schema.GroupKind) (clusterName string) {
	if a.GroupKind() != kind {
		return ""
	}
	return funcs.First(matcher(a.ClusterId(localId)))
}

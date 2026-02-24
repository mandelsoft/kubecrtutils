package owner_test

import (
	"context"
	"net/http"
	"net/url"

	"github.com/mandelsoft/kubecrtutils/enqueue"
	"github.com/mandelsoft/kubecrtutils/merge"
	"github.com/mandelsoft/kubecrtutils/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	sigclient "sigs.k8s.io/controller-runtime/pkg/client"
	sigcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"
)

type _cluster struct {
	name   string
	id     string
	scheme *runtime.Scheme
}

var _ types.Cluster = (*_cluster)(nil)

func (c *_cluster) LiftTechnical(clusterName string) (string, types.Cluster) {
	return c.name, c
}

func (c *_cluster) GetEventRecorder(name string) events.EventRecorder {
	// TODO implement me
	panic("implement me")
}

func (c *_cluster) GetAPIServerURL() (*url.URL, error) {
	// TODO implement me
	panic("implement me")
}

func (c *_cluster) WaitForCacheSync(ctx context.Context) bool {
	// TODO implement me
	panic("implement me")
}

func (_ *_cluster) TriggerSource(obj runtime.Object) (enqueue.TypedEnqueue[mcreconcile.Request], error) {
	// TODO implement me
	panic("implement me")
}

func (_ *_cluster) GetInfo() string {
	// TODO implement me
	panic("implement me")
}

func (_ *_cluster) GetTypeInfo() string {
	// TODO implement me
	panic("implement me")
}

func (_ *_cluster) IndexField(ctx context.Context, obj sigclient.Object, field string, extractValue sigclient.IndexerFunc) error {
	panic("implement me")
}

func (_ *_cluster) EnqueueByGVK(ctx context.Context, gvk schema.GroupVersionKind, key sigclient.ObjectKey) error {
	panic("implement me")
}

func (_ *_cluster) EnqueueByObject(ctx context.Context, obj runtime.Object) error {
	panic("implement me")
}

func (c *_cluster) GetName() string {
	return c.name
}

func (c *_cluster) GetId() string {
	return c.id
}

func (_ _cluster) GetScheme() *runtime.Scheme {
	panic("implement me")
}

func (c *_cluster) GetEffective() types.ClusterEquivalent {
	return c
}

func (c *_cluster) GetClusterById(id string) types.ClusterEquivalent {
	if c.id == id {
		return c
	}
	return nil
}

func (_ _cluster) CreateIndex(ctx context.Context, name string, proto sigclient.Object, indexer sigclient.IndexerFunc, wrap ...func(cluster types.ClusterEquivalent, name string) (types.Index, error)) (types.Index, error) {
	panic("implement me")
}

func (_ _cluster) Get(ctx context.Context, key sigclient.ObjectKey, obj sigclient.Object, opts ...sigclient.GetOption) error {
	panic("implement me")
}

func (_ _cluster) List(ctx context.Context, list sigclient.ObjectList, opts ...sigclient.ListOption) error {
	panic("implement me")
}

func (c *_cluster) AsCluster() types.Cluster {
	return c
}

func (_ _cluster) AsFleet() types.Fleet {
	return nil
}

func (c *_cluster) IsSameAs(o types.ClusterEquivalent) bool {
	return c.id == o.GetId()
}

func (c *_cluster) Filter(clusterName string, cluster sigcluster.Cluster) bool {
	return c.Match(clusterName)
}

func (c *_cluster) Match(clusterName string) bool {
	return c.GetName() == clusterName
}

func (c *_cluster) FilterById(clusterId string) bool {
	return c.id == clusterId
}

func (_ *_cluster) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...sigclient.ApplyOption) error {
	panic("implement me")
}

func (_ *_cluster) Create(ctx context.Context, obj sigclient.Object, opts ...sigclient.CreateOption) error {
	panic("implement me")
}

func (_ *_cluster) Delete(ctx context.Context, obj sigclient.Object, opts ...sigclient.DeleteOption) error {
	panic("implement me")
}

func (_ *_cluster) Update(ctx context.Context, obj sigclient.Object, opts ...sigclient.UpdateOption) error {
	panic("implement me")
}

func (_ *_cluster) Patch(ctx context.Context, obj sigclient.Object, patch sigclient.Patch, opts ...sigclient.PatchOption) error {
	panic("implement me")
}

func (_ *_cluster) DeleteAllOf(ctx context.Context, obj sigclient.Object, opts ...sigclient.DeleteAllOfOption) error {
	panic("implement me")
}

func (_ *_cluster) Status() sigclient.SubResourceWriter {
	panic("implement me")
}

func (_ *_cluster) SubResource(subResource string) sigclient.SubResourceClient {
	panic("implement me")
}

func (c *_cluster) Scheme() *runtime.Scheme {
	return c.scheme
}

func (_ *_cluster) RESTMapper() meta.RESTMapper {
	panic("implement me")
}

func (_ _cluster) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	panic("implement me")
}

func (_ *_cluster) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	panic("implement me")
}

func (_ *_cluster) GetHTTPClient() *http.Client {
	panic("implement me")
}

func (_ *_cluster) GetConfig() *rest.Config {
	panic("implement me")
}

func (_ *_cluster) GetCache() cache.Cache {
	panic("implement me")
}

func (_ *_cluster) GetClient() sigclient.Client {
	panic("implement me")
}

func (_ *_cluster) GetFieldIndexer() sigclient.FieldIndexer {
	panic("implement me")
}

func (_ *_cluster) GetEventRecorderFor(name string) record.EventRecorder {
	panic("implement me")
}

func (_ *_cluster) GetRESTMapper() meta.RESTMapper {
	panic("implement me")
}

func (_ *_cluster) GetAPIReader() sigclient.Reader {
	panic("implement me")
}

func (_ *_cluster) Start(ctx context.Context) error {
	panic("implement me")
}

func (c *_cluster) Unwrap() types.Cluster {
	return c
}

func (_ *_cluster) GetCluster() sigcluster.Cluster {
	panic("implement me")
}

func (_ *_cluster) GetIndex(name string) types.Index {
	panic("implement me")
}

func (_ *_cluster) GetTypeConverter() merge.Converters {
	panic("implement me")
}

func (_ *_cluster) ApplyTrigger(builder *ctrl.Builder, proto sigclient.Object) error {
	panic("implement me")
}

var _ types.Cluster = (*_cluster)(nil)

func NewCluster(clusterName, id string) types.Cluster {
	return &_cluster{
		name:   clusterName,
		id:     id,
		scheme: clientgoscheme.Scheme,
	}
}

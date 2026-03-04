/* Copyright 2025 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	runtime2 "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/multicluster-runtime/providers/clusters"
	"sigs.k8s.io/multicluster-runtime/providers/multi"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcbuilder "sigs.k8s.io/multicluster-runtime/pkg/builder"
	mcmanager "sigs.k8s.io/multicluster-runtime/pkg/manager"
	mcreconcile "sigs.k8s.io/multicluster-runtime/pkg/reconcile"

	apisv1alpha1 "github.com/kcp-dev/sdk/apis/apis/v1alpha1"
	corev1alpha1 "github.com/kcp-dev/sdk/apis/core/v1alpha1"
	tenancyv1alpha1 "github.com/kcp-dev/sdk/apis/tenancy/v1alpha1"

	"github.com/kcp-dev/multicluster-provider/apiexport"

	corednsv1alpha1 "github.com/mandelsoft/kubedns/api/coredns/v1alpha1"
)

var (
	myscheme = runtime2.NewScheme()
)

func init() {
	runtime.Must(corev1alpha1.AddToScheme(myscheme))
	runtime.Must(tenancyv1alpha1.AddToScheme(myscheme))
	runtime.Must(apisv1alpha1.AddToScheme(myscheme))

	runtime.Must(clientgoscheme.AddToScheme(myscheme))

	runtime.Must(corednsv1alpha1.AddToScheme(myscheme))
}

const ANNOTATION = "mandelsoft.org/owner"

func main() {
	log.SetLogger(zap.New(zap.UseDevMode(true)))

	ctx := signals.SetupSignalHandler()
	entryLog := log.Log.WithName("entrypoint")

	var (
		kcp           string
		endpointSlice string
		runtime       string
	)

	pflag.StringVar(&kcp, "kubeconfig", "", "Set KCP workspace kubeconfig for the api provider")
	pflag.StringVar(&endpointSlice, "endpointslice", "examples-apiexport-multicluster", "Set the APIExportEndpointSlice name to watch")
	pflag.StringVar(&runtime, "runtime", "", "Set runtime kubeconfig used to implement the apir")
	pflag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags("", kcp)
	if err != nil {
		entryLog.Error(err, "get rest config for kcp provider")
		os.Exit(1)
	}

	rtcfg, err := clientcmd.BuildConfigFromFlags("", runtime)
	if err != nil {
		entryLog.Error(err, "get rest config for runtime cluster")
		os.Exit(1)
	}

	// Setup a Manager, note that this not yet engages clusters, only makes them available.
	entryLog.Info("Setting up manager")
	opts := manager.Options{
		Scheme: myscheme,
	}

	kcpprovider, err := apiexport.New(cfg, endpointSlice, apiexport.Options{
		Scheme: myscheme,
	})
	if err != nil {
		entryLog.Error(err, "unable to construct cluster provider")
		os.Exit(1)
	}

	clusterprovider := clusters.New()
	rt, err := cluster.New(rtcfg)
	if err != nil {
		entryLog.Error(err, "unable to set up runtime cluster")
		os.Exit(1)
	}
	clusterprovider.Add(ctx, "runtime", rt)

	multiprovider := multi.New(multi.Options{})
	if err != nil {
		entryLog.Error(err, "unable to add cluster provider")
		os.Exit(1)
	}
	err = multiprovider.AddProvider("", clusterprovider)
	if err != nil {
		entryLog.Error(err, "unable to add cluster provider")
		os.Exit(1)
	}

	err = multiprovider.AddProvider("dataplane", kcpprovider)
	if err != nil {
		entryLog.Error(err, "unable to add kcp provider")
		os.Exit(1)
	}

	mgr, err := mcmanager.New(cfg, multiprovider, opts)
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	if err := kcpprovider.IndexField(
		ctx,
		&corednsv1alpha1.HostedZone{},
		"IndexKeyZoneParent",
		func(obj client.Object) []string {
			r := obj.(*corednsv1alpha1.HostedZone)
			if r.Spec.ParentRef == "" {
				return nil
			}
			return []string{r.Spec.ParentRef}
		},
	); err != nil {
		entryLog.Error(err, "unable to set up index")
		os.Exit(1)
	}

	if err := mcbuilder.ControllerManagedBy(mgr).
		Named("kcp-test-controller").
		For(&corednsv1alpha1.HostedZone{},
			mcbuilder.WithClusterFilter(
				func(clusterName string, cluster cluster.Cluster) bool {
					return strings.HasPrefix(clusterName, "dataplane#")
				}),
		).
		/*
			Watches(&v1.Secret{}, HandlerFactory(mgr.GetLogger()), mcbuilder.WithClusterFilter(
				func(clusterName string, cluster cluster.Cluster) bool {
					return clusterName == "#runtime"
				}),
			).

		*/
		Complete(mcreconcile.Func(
			func(ctx context.Context, req mcreconcile.Request) (ctrl.Result, error) {
				log := log.FromContext(ctx).WithValues("cluster", req.ClusterName)

				log.Info("*** Request", "request", req)
				cl, err := mgr.GetCluster(ctx, req.ClusterName)
				if err != nil {
					return reconcile.Result{}, fmt.Errorf("failed to get cluster: %w", err)
				}
				clt := cl.GetClient()

				// Retrieve the ConfigMap from the cluster.
				s := &corednsv1alpha1.HostedZone{}
				if err := clt.Get(ctx, req.NamespacedName, s); err != nil {
					if apierrors.IsNotFound(err) {
						log.Info("Zone deleted", "name", s.Name, "uuid", s.UID)
						return reconcile.Result{}, nil
					}
					return reconcile.Result{}, fmt.Errorf("failed to get configmap: %w", err)
				}

				log.Info("Reconciling Zone", "name", s.Name, "uuid", s.UID)
				// recorder := cl.GetEventRecorderFor("kcp-configmap-controller")
				// recorder.Eventf(s, corev1.EventTypeNormal, "Zone Reconciled", "Zone %s reconciled", s.Name)

				err = cl.GetCache().List(ctx, &corednsv1alpha1.HostedZoneList{}, client.MatchingFields{"IndexKeyZoneParent": req.Name})
				if err != nil {
					log.Info("=== cluster index error: {{error}}", "error", err)
				} else {
					log.Info("=== cluster index found}")
				}

				var secret v1.Secret

				err = cl.GetCache().Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: "dns-service"}, &secret)
				if err == nil {
					log.Info("string(ca.crt)", "cert", string(secret.Data["ca.crt"]))
					log.Info("ca.crt", "cert", secret.Data["ca.crt"])
				}
				return reconcile.Result{}, nil
			},
		)); err != nil {
		entryLog.Error(err, "failed to build controller")
		os.Exit(1)
	}

	entryLog.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}

func HandlerFactory(log logr.Logger) func(clusterName string, _ cluster.Cluster) handler.TypedEventHandler[client.Object, mcreconcile.Request] {
	return func(clusterName string, _ cluster.Cluster) handler.TypedEventHandler[client.Object, mcreconcile.Request] {
		return &_handler{log, clusterName}
	}
}

type _handler struct {
	log         logr.Logger
	clusterName string
}

var _ handler.TypedEventHandler[client.Object, mcreconcile.Request] = (*_handler)(nil)

func (h *_handler) Create(ctx context.Context, e event.TypedCreateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.trigger(w, e.Object)
}

func (h *_handler) Update(ctx context.Context, e event.TypedUpdateEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.trigger(w, e.ObjectNew)
}

func (h *_handler) Delete(ctx context.Context, e event.TypedDeleteEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.trigger(w, e.Object)
}

func (h *_handler) Generic(ctx context.Context, e event.TypedGenericEvent[client.Object], w workqueue.TypedRateLimitingInterface[mcreconcile.Request]) {
	h.trigger(w, e.Object)
}

func (h *_handler) trigger(w workqueue.TypedRateLimitingInterface[mcreconcile.Request], obj client.Object) {
	annos := obj.GetAnnotations()
	if annos == nil {
		return
	}
	v := annos[ANNOTATION]
	if v == "" {
		return
	}
	fields := strings.Split(v, "/")
	if len(fields) != 3 {
		return
	}
	h.log.Info("TRIGGER", "key", v)
	w.Add(mcreconcile.Request{ClusterName: fields[0], Request: reconcile.Request{types.NamespacedName{Namespace: fields[1], Name: fields[2]}}})
}

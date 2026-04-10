package main

import (
	"os"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/ctrlmgmt"
	"github.com/mandelsoft/kubecrtutils/examples/simple/controllers"
	"github.com/mandelsoft/kubecrtutils/examples/simple/controllers/replicate"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	"github.com/mandelsoft/kubecrtutils/options/metricsopts"
	"github.com/mandelsoft/kubecrtutils/options/mlogopts"
	"github.com/mandelsoft/kubecrtutils/setup"
	"github.com/spf13/pflag"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// +kubebuilder:scaffold:imports
)

// --- begin scheme ---
var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// --- end scheme ---

func main() {

	// --- begin orchestrate controller manager ---

	def := ctrlmgmt.Define("replicator.mandelsoft.de", controllers.SOURCE).
		WithScheme(scheme).
		AddCluster(
			cluster.Define(controllers.SOURCE, "replication source").WithFallback(cluster.DEFAULT),
			cluster.Define(controllers.TARGET, "replication target").WithFallback(controllers.SOURCE),
		).
		AddController(
			replicate.Controller(),
		)
	// --- end orchestrate controller manager ---

	// --- begin orchestrate general functionality ---
	options := &flagutils.DefaultOptionSet{}

	options.Add(
		metricsopts.New(),    // options to control the manager metrics service
		mlogopts.New(true),   // options to control mandelsoft/logging
		activationopts.New(), // enable controller selection
		// other options
	)
	// --- end orchestrate general functionality ---

	// --- begin execute everything ---
	err := ctrlmgmt.Setup("replicator", options, def, os.Args[1:]...)
	if err == pflag.ErrHelp {
		os.Exit(0)
	}
	setup.ExitIfErr(err, "setup controller manager")
	// --- end execute everything ---
}

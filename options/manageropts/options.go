package manageropts

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/kubecrtutils/cluster"
	"github.com/mandelsoft/kubecrtutils/options/metricsopts"
	"github.com/mandelsoft/kubecrtutils/options/tlsopts"
	"github.com/mandelsoft/kubecrtutils/options/webhookopts"
	"github.com/mandelsoft/logging"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrl "sigs.k8s.io/multicluster-runtime"
	"sigs.k8s.io/multicluster-runtime/providers/clusters"
	"sigs.k8s.io/multicluster-runtime/providers/multi"
)

type Options struct {
	// main is the cluster used as main cluster for the manager
	main                    string
	Nested                  flagutils.OptionSet
	EnableLeaderElection    bool
	LeaderElectionNamespace string
	ProbeAddr               string
	ElectionId              string

	defaultElectionId string

	// Configurations describes a sequence of ConfigurationProvider.
	// They are used to finalize the manager options before
	// the manager is created.
	Configurations []ConfigurationProvider
}

func From(opts flagutils.OptionSetProvider) *Options {
	return flagutils.GetFrom[*Options](opts)
}

var (
	_ flagutils.Options     = (*Options)(nil)
	_ flagutils.Validatable = (*Options)(nil)
)

func New(main string, electionId string, configs ...ConfigurationProvider) *Options {
	if main == "" {
		main = cluster.DEFAULT
	}
	nested := flagutils.NewOptionSet(tlsopts.New())
	return &Options{Nested: nested, defaultElectionId: electionId, Configurations: configs, main: main}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ElectionId, "leader-election-id", o.ElectionId, "Id for leader election")
	fs.StringVar(&o.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.StringVar(&o.LeaderElectionNamespace, "leader-elect-namespace", "", "leader election namespace")
	fs.BoolVar(&o.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	o.Nested.AddFlags(fs)
}

func (o *Options) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	err := flagutils.Validate(ctx, o.Nested, v)
	if err != nil {
		return err
	}

	clusters, err := cluster.ValidatedClusters(ctx, opts, v)
	if err != nil {
		return err
	}

	main := clusters.Get(o.main)
	if main == nil {
		return fmt.Errorf("could not find main cluster %q", o.main)
	}

	_, err = flagutils.ValidatedOptions[*metricsopts.Options](ctx, opts, v)
	if err != nil {
		return err
	}

	_, err = flagutils.ValidatedOptions[*webhookopts.Options](ctx, opts, v)
	if err != nil {
		return err
	}

	_, err = flagutils.ValidatedFilteredOptions[ConfigurationProvider](ctx, opts, v)
	return err
}

// AsOptionSet provides access o the nested option set.
func (o *Options) AsOptionSet() flagutils.OptionSet {
	return o.Nested
}

////////////////////////////////////////////////////////////////////////////////

func (o *Options) GetMain() string {
	return o.main
}

func (o *Options) GetManager(ctx context.Context, opts flagutils.OptionSetProvider) (ctrl.Manager, error) {
	configuredClusters := cluster.From(opts).GetClusters()
	if configuredClusters == nil {
		return nil, fmt.Errorf("no cluster definitions found in options")
	}

	cl := configuredClusters.Get(o.main)
	if cl == nil {
		return nil, fmt.Errorf("could not find main cluster %q", o.main)
	}

	var main cluster.Cluster
	if cl.AsFleet() != nil {
		if cl.AsFleet().GetBaseCluster() == nil {
			return nil, fmt.Errorf("could not find base cluster for fleet %q to instantiate manager", o.main)
		}
		main = cl.AsFleet().GetBaseCluster()
	} else {
		main = cl.AsCluster()
	}

	metrics := metricsopts.From(opts)
	web := webhookopts.From(opts)

	configs := flagutils.Filter[ConfigurationProvider](opts)

	cfg := ctrl.Options{
		Logger:                  logging.DefaultContext().Logger(logging.NewRealm("controller-manager")).V(4),
		Scheme:                  main.GetScheme(),
		Metrics:                 metrics.GetMetricsServerOpts(),
		WebhookServer:           web.GetServer(),
		HealthProbeBindAddress:  o.ProbeAddr,
		LeaderElection:          o.EnableLeaderElection,
		LeaderElectionNamespace: o.LeaderElectionNamespace,
		LeaderElectionID:        o.defaultElectionId,
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,

		// implicit cluster creation cannot be circumvented (why), so fake
		// using shared info as far as possible.,
		NewClient: func(config *rest.Config, options client.Options) (client.Client, error) {
			return main.GetClient(), nil
		},
		NewCache: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
			return main.GetCache(), nil
		},
	}

	for _, conf := range configs {
		err := conf.Configure(ctx, &cfg, opts.AsOptionSet())
		if err != nil {
			return nil, err
		}
	}

	if o.ElectionId != "" {
		cfg.LeaderElectionID = o.ElectionId
	}

	for _, conf := range o.Configurations {
		err := conf.Configure(ctx, &cfg, opts.AsOptionSet())
		if err != nil {
			return nil, err
		}
	}

	provider := multi.New(multi.Options{Separator: "#", ChannelSize: configuredClusters.Len()})
	m, err := ctrl.NewManager(main.GetConfig(), provider, cfg)
	if err != nil {
		return nil, err
	}

	clusterprovider := clusters.New()
	provider.AddProvider("", clusterprovider)

	if err := m.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := m.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("unable to set up ready check: %w", err)
	}

	found := sets.New[string]()
	for _, c := range configuredClusters.Elements {
		eff := c.GetEffective()
		if !found.Has(c.GetName()) {
			found.Insert(c.GetName())
			if eff.AsFleet() != nil {
				cfg.Logger.Info("adding fleet {{fleet}} -> {{effective}}", "fleet", c.GetName(), "effective", c.GetEffective().GetName())
				err = provider.AddProvider(c.GetName(), c.AsFleet().GetProvider())
			} else {
				cfg.Logger.Info("adding cluster {{cluster}} -> {{effective}}", "cluster", c.GetName(), "effective", c.GetEffective().GetName())
				err = clusterprovider.Add(ctx, eff.GetName(), eff.AsCluster())
			}
		} else {
			cfg.Logger.Info("cluster {{cluster}} -> {{effective}} already added", "cluster", c.GetName(), "effective", c.GetEffective().GetName())
		}
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

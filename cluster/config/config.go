package config

import (
	"fmt"
	"os"

	"github.com/mandelsoft/goutils/general"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type SubConfig interface {
	GetType() string
}

type ConfigOptions struct {
	clientcmd.ConfigOverrides
	Identity  string
	SubConfig SubConfig
}

type Config struct {
	Identity   string
	RestConfig *rest.Config
	Namespace  string
	Context    string
	SubConfig  SubConfig
}

func (c *Config) GetId() string {
	if c.Identity == "" && c.Context != "" {
		// Good idea to use the context as Id ???
		// return ropts.CurrentContext
	}
	return c.Identity
}

////////////////////////////////////////////////////////////////////////////////

func PrepareConfigFor(opts *ConfigOptions) *Config {
	return &Config{
		Identity:  opts.Identity,
		Context:   opts.ConfigOverrides.CurrentContext,
		Namespace: opts.ConfigOverrides.Context.Namespace,
		SubConfig: opts.SubConfig,
	}
}

// GetConfig provides a rest config based on an optional path
// (for a kubeconfig file) and an optional context.
// If path is in-cluster the usage of an in-cluster kubeconfig
// is enforced.
// If the path is empty the usual suspects as defined in
// clientcmd.NewDefaultClientConfigLoadingRules are used
// falling back to the in-cluster kubeconfig.
// If a path is given, it must provide a kubeconfig.
func GetConfig(path string, context string) (*Config, error) {
	if path == "in-cluster" {
		rest, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return &Config{RestConfig: rest}, nil
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = path
	cfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules,
		&clientcmd.ConfigOverrides{CurrentContext: context})

	rest, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	ns, _, err := cfg.Namespace()
	if err != nil {
		return nil, err
	}

	merged, err := cfg.RawConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		RestConfig: rest,
		Namespace:  ns,
		Context:    merged.CurrentContext,
	}, nil
}

func GetRestConfigFromKubeconfig(apiConfig *api.Config, overrides ...*ConfigOptions) (*Config, error) {
	// 1. Create a ClientConfig object from the api.Config
	// We use NewDefaultClientConfig to respect the "current-context" in the struct.
	over := general.OptionalDefaulted(&ConfigOptions{}, overrides...)
	clientConfig := clientcmd.NewDefaultClientConfig(*apiConfig, &over.ConfigOverrides)

	merged, err := clientConfig.MergedRawConfig()
	if err != nil {
		return nil, err
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	ns, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, err
	}

	cfg := PrepareConfigFor(over)
	cfg.RestConfig = restConfig
	cfg.Context = merged.CurrentContext
	cfg.Namespace = ns
	return cfg, nil
}

func TryKubeconfigFile(path string, opts *ConfigOptions) (*Config, error) {
	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("kubeconfig file %s: %w", path, err)
	}
	return GetRestConfigFromKubeconfig(config, opts)
}

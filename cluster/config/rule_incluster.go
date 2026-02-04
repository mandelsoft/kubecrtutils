package config

import (
	"fmt"
	"os"

	"k8s.io/client-go/rest"
)

type InClusterConfig struct{}

func NewInClusterConfig() Rule {
	return InClusterConfig{}
}

func (r InClusterConfig) GetConfig(opts *ConfigOptions) (*Config, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		if err == os.ErrNotExist || err == rest.ErrNotInCluster {
			return nil, nil
		}
		return nil, err
	}
	if opts != nil && opts.SubConfig != nil {
		return nil, fmt.Errorf("incluster config not possible for sub type")
	}
	return &Config{RestConfig: cfg, Identity: opts.Identity, Context: opts.CurrentContext}, nil
}

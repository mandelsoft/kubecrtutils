package config

import (
	"os"
	"strings"

	"github.com/mandelsoft/goutils/general"
)

type EnvironmentVariable struct {
	Name string
}

func NewEnvironmentVariable(name ...string) *EnvironmentVariable {
	return &EnvironmentVariable{general.OptionalDefaulted("KUBECONFIG", name...)}
}

func (r *EnvironmentVariable) PersonalizedWith(o *Personalization) Rule {
	name := o.Name
	if len(name) == 0 {
		name = r.Name
	} else {
		name = strings.ToUpper(name) + "_" + r.Name
	}
	return &EnvironmentVariable{
		Name: name,
	}
}

func (r *EnvironmentVariable) GetConfig(opts *ConfigOptions) (*Config, error) {
	v := os.Getenv(r.Name)
	if v == "" {
		return nil, nil
	}
	cfg, err := TryKubeconfigFile(v, opts)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

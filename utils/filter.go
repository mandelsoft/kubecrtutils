package utils

import (
	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/reflectutils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
)

func GetUsed[T any, S set.Set[string]](opts flagutils.OptionSet) S {
	filters := flagutils.Filter[T](opts)
	required := set.New[string]()

	copts := activationopts.From(opts)
	if copts != nil {
		for _, f := range filters {
			required.AddAll(reflectutils.CallMethodByInterfaceR[T, set.Set[string]](f, copts.GetContraintContext()))
		}
	}
	return S(required)
}

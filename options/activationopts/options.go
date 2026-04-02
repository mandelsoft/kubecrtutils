package activationopts

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
	"github.com/mandelsoft/kubecrtutils/types"
	"github.com/spf13/pflag"
)

const ALL = "all"

type ControllerSet = constraints.ControllerSet

type ControllerSource interface {
	GetControllerSet() ControllerSet
}

func From(set flagutils.OptionSet) *Options {
	return flagutils.GetFrom[*Options](set)
}

type Options struct {
	set       ControllerSet
	names     []string
	activated []string
}

var (
	_ flagutils.Options     = (*Options)(nil)
	_ flagutils.Preparable  = (*Options)(nil)
	_ flagutils.Validatable = (*Options)(nil)
)

// New provides options usable to activate or deactivate controllers
func New() *Options {
	return &Options{}
}

func (o *Options) Prepare(ctx context.Context, opts flagutils.OptionSet, v flagutils.PreparationSet) error {
	// check controller definitions (without preparation)
	cdefs := flagutils.GetFrom[ControllerSource](opts)
	if cdefs != nil {
		o.set = cdefs.GetControllerSet()
		names := o.set.GetNames()
		if len(names) > 0 {
			names = sliceutils.CopyAppend(names, ALL)
			names = append(names, maputils.OrderedKeys(o.set.GetGroups())...)
			o.names = names
		}
	}
	return nil
}

func (o *Options) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	var invalid []string
	for _, name := range o.activated {
		name = strings.TrimLeft(name, "+-")
		if !slices.Contains(o.names, name) {
			invalid = append(invalid, name)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("invalid controller names or groups: %v", invalid)
	}
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	if len(o.names) != 0 {
		fs.StringSliceVarP(&o.activated, "controllers", "", []string{ALL}, fmt.Sprintf("activated controllers (%s).", strings.Join(o.names, ", ")))
	}
}

////////////////////////////////////////////////////////////////////////////////

func (o *Options) handle(h func(name ...string) set.Set[string], name string, handled set.Set[string], simplified map[string]string) {
	var list []string
	if n, ok := simplified[name]; ok {
		name = n
	}
	if handled.Contains(name) {
		return
	}
	handled.Add(name)
	grps := o.set.GetGroups()
	if grps != nil {
		if l, ok := grps[name]; ok {
			list = l
		}
	}
	if list == nil {
		if name == ALL {
			list = o.set.GetNames()
		}
	}
	if list != nil {
		for _, name := range list {
			o.handle(h, name, handled, simplified)
		}
	} else {
		h(name)
	}
	handled.Delete(name)
}

func (o *Options) GetContraintContext() *constraints.Context {
	return constraints.NewContext(o.set).WithSelectedSet(o.GetActivation())
}

func (o *Options) GetActivation() types.ControllerNames {
	simplified := Simplify(o.names)
	handled := types.ControllerNames{}
	names := set.New[string]()
	for i, name := range o.activated {
		if strings.HasPrefix(name, "-") {
			if i == 0 {
				names.Add(o.set.GetNames()...)
			}
			o.handle(names.Delete, name[1:], handled, simplified)
		} else {
			if strings.HasPrefix(name, "+") {
				o.handle(names.Add, name[1:], handled, simplified)
			} else {
				o.handle(names.Add, name, handled, simplified)
			}
		}

	}
	return names
}

func Simplify(set []string) map[string]string {
	result := map[string]string{}

next:
	for _, n := range set {
		comps := strings.Split(n, ".")
		if len(comps) == 1 {
			result[n] = n
		} else {
			for i := range comps {
				t := strings.Join(comps[len(comps)-i-1:], ".")
				if result[t] == "" {
					result[t] = n
					continue next
				}
			}
		}
	}
	return result
}

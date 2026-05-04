package controllerutils

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/mandelsoft/flagutils"
	"github.com/spf13/pflag"
	"golang.org/x/exp/maps"
)

type None struct{}

type DefaultElement interface {
	IsDefault() bool
}

// Factory acts a factory for a target object using a configuration.
type Factory[C any, O any] interface {
	Create(ctx context.Context, cfg C) (O, error)
	Description() string
}

type FactoryFunc[C any, O any] func(context.Context, C) (O, error)

func (f FactoryFunc[C, O]) Create(ctx context.Context, cfg C) (O, error) {
	return f(ctx, cfg)
}

func (f FactoryFunc[C, O]) Description() string {
	return ""
}

// Registry registers a set of named factory alternative for a particular purpose.
// It can then be used to create an appropriate instace for a given type name and configuration.
type Registry[C any, O any] interface {
	flagutils.OptionSet

	Clone() Registry[C, O]
	SelectedMode() string

	Register(name string, factory Factory[C, O])
	Names() []string
	Description() string

	Get(name string) Factory[C, O]
	Create(ctx context.Context, name string, cfg C) (O, error)

	CreateConfigured(ctx context.Context, opts C) (O, error)
}

type registry[C, O any] struct {
	lock      sync.RWMutex
	typ       string
	desc      string
	def       string
	factories map[string]Factory[C, O]
	selected  string

	flagutils.DefaultOptionSet
}

func NewRegistry[C any, O any](typ, desc string) Registry[C, O] {
	return &registry[C, O]{typ: typ, factories: make(map[string]Factory[C, O]), desc: desc}
}

func (r *registry[C, O]) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&r.selected, r.typ, "", r.def, r.desc+"("+strings.Join(r.Names(), ", ")+")")
	r.DefaultOptionSet.AddFlags(fs)
}

func (r *registry[C, O]) SelectedMode() string {
	return r.selected
}

func (r *registry[C, O]) Clone() Registry[C, O] {
	r.lock.Lock()
	defer r.lock.Unlock()
	n := NewRegistry[C, O](r.typ, r.desc)
	for k, v := range r.factories {
		n.Register(k, v)
	}
	return n
}

func (r *registry[C, O]) Description() string {
	return r.desc
}

func (r *registry[C, O]) Register(name string, f Factory[C, O]) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.factories[name] = f
	if o, ok := f.(flagutils.Options); ok {
		r.DefaultOptionSet.Add(o)
	}
	if d, ok := f.(DefaultElement); ok {
		if d.IsDefault() {
			r.def = name
			r.selected = name
		}
	}
}

func (r *registry[C, O]) Names() []string {
	r.lock.Lock()
	defer r.lock.Unlock()
	keys := maps.Keys(r.factories)
	sort.Strings(keys)
	return keys
}

func (r *registry[C, O]) Get(name string) Factory[C, O] {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.factories[name]
}

func (r *registry[C, O]) CreateConfigured(ctx context.Context, opts C) (O, error) {
	var _nil O
	if r.selected == "" {
		return _nil, fmt.Errorf("no %s configured", r.typ)
	}
	return r.Create(ctx, r.selected, opts)
}

func (r *registry[C, O]) Create(ctx context.Context, name string, cfg C) (O, error) {
	var _nil O

	r.lock.Lock()
	defer r.lock.Unlock()

	f := r.factories[name]
	if f == nil {
		return _nil, fmt.Errorf("unknown %s type %q", r.typ, name)
	}
	return f.Create(ctx, cfg)
}

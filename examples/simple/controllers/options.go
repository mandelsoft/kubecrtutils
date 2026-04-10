package controllers

import (
	"context"
	"fmt"

	"github.com/mandelsoft/flagutils"
	"github.com/spf13/pflag"
)

type Options struct {
	Annotation string
	Class      string
}

var _ flagutils.Validatable = (*Options)(nil)

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Annotation, "annotation", CLASS_ANNOTATION, "annotation name holding the class")
	fs.StringVar(&o.Class, "class", "", "replication class")
}

func (o *Options) Validate(ctx context.Context, opts flagutils.OptionSet, v flagutils.ValidationSet) error {
	if o.Annotation == "" {
		return fmt.Errorf("annotation name is required")
	}
	return nil
}

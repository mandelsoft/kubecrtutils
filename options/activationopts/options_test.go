package activationopts_test

import (
	"context"

	"github.com/mandelsoft/flagutils"
	"github.com/mandelsoft/goutils/maputils"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

type cset struct {
	groups map[string][]string
	names  []string
}

var (
	_ flagutils.Options               = (*cset)(nil)
	_ activationopts.ControllerSet    = (*cset)(nil)
	_ activationopts.ControllerSource = (*cset)(nil)
)

func (c *cset) AddFlags(fs *pflag.FlagSet) {
}

func (c *cset) GetControllerSet() activationopts.ControllerSet {
	return c
}

func (c *cset) GetNames() []string {
	return c.names
}

func (c *cset) GetGroups() map[string][]string {
	return c.groups
}

var _ = Describe("Calculation", func() {
	var flags *pflag.FlagSet
	var s *cset
	var opts flagutils.DefaultOptionSet
	var myopt *activationopts.Options

	BeforeEach(func() {
		flags = pflag.NewFlagSet("test", pflag.ExitOnError)
		s = &cset{
			groups: map[string][]string{
				"A": []string{"a", "b"},
				"B": []string{"c", "d"},
				"C": []string{"A", "c"},
				"D": []string{"A", "D"}, // recursion
			},
			names: []string{"a", "b", "c", "d", "e"},
		}
		myopt = activationopts.New()
		opts = nil
		opts.Add(s)
		opts.Add(myopt)
		MustBeSuccessful(flagutils.Prepare(context.Background(), opts, nil))
		opts.AddFlags(flags)
	})

	Context("direct", func() {

		It("default", func() {
			MustBeSuccessful(flags.Parse([]string{}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"a", "b", "c", "d", "e"}))
		})

		It("all", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=all"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"a", "b", "c", "d", "e"}))
		})

		It("positive list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=+a,b", "--controllers=c"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"a", "b", "c"}))
		})

		It("negative list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=-a,-b", "--controllers=-c"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"d", "e"}))
		})

		It("mixed positive list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=+a,+b", "--controllers=-c,-b"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"a"}))
		})

		It("mixed negative list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=-a,-b", "--controllers=-c,+b"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"b", "d", "e"}))
		})
	})

	Context("groups", func() {
		It("positive list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=+A,+B"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"a", "b", "c", "d"}))
		})
		It("negative list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=-A,-B"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"e"}))
		})
		It("mixed positive list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=C,-A"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"c"}))
		})
		It("mixed negative list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=-C,A"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"a", "b", "d", "e"}))
		})

		It("cyclic positive list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=D"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"a", "b"}))
		})

		It("cyclic negative list", func() {
			MustBeSuccessful(flags.Parse([]string{"--controllers=-D"}))
			MustBeSuccessful(flagutils.Validate(context.Background(), opts, nil))
			Expect(maputils.OrderedKeys(myopt.GetActivation())).To(Equal([]string{"c", "d", "e"}))
		})
	})
})

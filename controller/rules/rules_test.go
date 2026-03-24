package rules_test

import (
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/kubecrtutils/controller/rules"
	"github.com/mandelsoft/kubecrtutils/options/activationopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type setting struct {
	groups map[string][]string
	names  []string
}

var _ activationopts.ControllerSet = (*setting)(nil)

func (s *setting) GetNames() []string {
	return s.names
}

func (s *setting) GetGroups() map[string][]string {
	return s.groups
}

var _ = Describe("Test Environment", func() {
	Context("", func() {
		cfg := &setting{
			names: []string{"a", "b", "c", "d", "e"},
			groups: map[string][]string{
				"A": {"a", "b"},
				"B": {"c", "d"},
				"C": {"A", "c"},
				"D": {"A", "D"},
			},
		}
		ctx := rules.NewContext(cfg)

		It("creates context", func() {
			Expect(ctx.Names().AsArray()).To(ConsistOf("a", "b", "c", "d", "e"))
		})

		It("creates closures", func() {
			Expect(ctx.GetGroup("A").AsArray()).To(ConsistOf("a", "b"))
			Expect(ctx.GetGroup("B").AsArray()).To(ConsistOf("c", "d"))
			Expect(ctx.GetGroup("C").AsArray()).To(ConsistOf("a", "b", "c"))
			Expect(ctx.GetGroup("D").AsArray()).To(ConsistOf("a", "b"))
		})

		Context("complete", func() {
			ruleset := rules.New().Add(rules.Complete("A"))
			It("succeeds", func() {
				cur := set.New[string]("a", "b", "c")
				testutils.MustBeSuccessful(ruleset.Match(ctx, cur))
			})
			It("failes", func() {
				cur := set.New[string]("a")
				Expect(ruleset.Match(ctx, cur)).To(MatchError("group \"A\" must be complete [a, b]"))
			})

			It("succeeds for one rule with two groups", func() {
				ruleset := rules.New().Add(rules.Complete("A", "B"))
				cur := set.New[string]("a", "b", "c", "d")
				testutils.MustBeSuccessful(ruleset.Match(ctx, cur))
			})
			It("succeeds for two rules", func() {
				ruleset := rules.New().Add(rules.Complete("A")).Add(rules.Complete("B"))
				cur := set.New[string]("a", "b", "c", "d")
				testutils.MustBeSuccessful(ruleset.Match(ctx, cur))
			})

			It("fails for one rule with two groups", func() {
				ruleset := rules.New().Add(rules.Complete("A", "B"))
				cur := set.New[string]("a", "b", "d")
				Expect(ruleset.Match(ctx, cur)).To(MatchError("group \"B\" must be complete [c, d]"))
			})
			It("succeeds for two rules", func() {
				ruleset := rules.New().Add(rules.Complete("A")).Add(rules.Complete("B"))
				cur := set.New[string]("a", "b", "d")
				Expect(ruleset.Match(ctx, cur)).To(MatchError("group \"B\" must be complete [c, d]"))
			})
		})

		Context("disjoint", func() {
			ruleset := rules.New().Add(rules.Disjoint("A", "B"))
			It("succeeds", func() {
				cur := set.New[string]("a", "e")
				testutils.MustBeSuccessful(ruleset.Match(ctx, cur))
			})

			It("failes", func() {
				cur := set.New[string]("a", "b", "c")
				Expect(ruleset.Match(ctx, cur)).To(MatchError("use only controllers either in group \"A\" or \"B\""))
			})
		})
	})
})

package constraints_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/kubecrtutils/controller/constraints"
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
		ctx := constraints.NewContext(cfg)

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
			constraintset := constraints.New().Add(constraints.Complete("A"))
			It("succeeds", func() {
				MustBeSuccessful(constraintset.Match(ctx.WithSelected("a", "b", "c")))
			})
			It("failes", func() {
				ExpectError(constraintset.Match(ctx.WithSelected("a"))).To(MatchError("group \"A\" must be complete [a, b]"))
			})

			It("succeeds for one constraint with two groups", func() {
				constraintset := constraints.New().Add(constraints.Complete("A", "B"))
				MustBeSuccessful(constraintset.Match(ctx.WithSelected("a", "b", "c", "d")))
			})
			It("succeeds for two constraints", func() {
				constraintset := constraints.New().Add(constraints.Complete("A")).Add(constraints.Complete("B"))
				MustBeSuccessful(constraintset.Match(ctx.WithSelected("a", "b", "c", "d")))
			})

			It("fails for one constraint with two groups", func() {
				constraintset := constraints.New().Add(constraints.Complete("A", "B"))
				ExpectError(constraintset.Match(ctx.WithSelected("a", "b", "d"))).To(MatchError("group \"B\" must be complete [c, d]"))
			})
			It("succeeds for two constraints", func() {
				constraintset := constraints.New().Add(constraints.Complete("A")).Add(constraints.Complete("B"))
				ExpectError(constraintset.Match(ctx.WithSelected("a", "b", "d"))).To(MatchError("group \"B\" must be complete [c, d]"))
			})
		})

		Context("disjoint", func() {
			constraintset := constraints.New().Add(constraints.Disjoint("A", "B"))
			It("succeeds", func() {
				MustBeSuccessful(constraintset.Match(ctx.WithSelected("a", "e")))
			})

			It("failes", func() {
				ExpectError(constraintset.Match(ctx.WithSelected("a", "b", "c"))).To(MatchError("use only controllers either in group \"A\" or \"B\""))
			})
		})

		Context("activated", func() {
			constraintset := constraints.New().Add(constraints.Activated("c"))
			It("succeeds", func() {
				Expect(constraintset.Match(ctx.WithSelected("c", "d"))).To(Equal(constraints.Yes))
			})

			It("fails", func() {
				Expect(constraintset.Match(ctx.WithSelected("d"))).To(Equal(constraints.No))
			})
		})

		Context("operations", func() {
			Context("or", func() {
				constraintset := constraints.New().Add(constraints.Or(constraints.Complete("A"), constraints.Complete("B")))

				It("succeeds on empty", func() {
					constraintset := constraints.New().Add(constraints.Or())
					Expect(constraintset.Match(ctx.WithSelected("a", "b", "c", "d"))).To(Equal(constraints.NoOpinion))
				})

				It("succeeds", func() {
					MustBeSuccessful(constraintset.Match(ctx.WithSelected("c", "d")))
				})

				It("fails", func() {
					ExpectError(constraintset.Match(ctx.WithSelected("c", "e"))).To(MatchError("group \"A\" must be complete [a, b] OR group \"B\" must be complete [c, d]"))
				})
			})

			Context("and", func() {
				constraintset := constraints.New().Add(constraints.And(constraints.Complete("A"), constraints.Complete("B")))

				It("succeeds on empty", func() {
					constraintset := constraints.New()
					Expect(constraintset.Match(ctx.WithSelected("a", "b", "c", "d"))).To(Equal(constraints.NoOpinion))
				})

				It("succeeds", func() {
					MustBeSuccessful(constraintset.Match(ctx.WithSelected("a", "b", "c", "d")))
				})

				It("fails one", func() {
					ExpectError(constraintset.Match(ctx.WithSelected("a", "b"))).To(MatchError("group \"B\" must be complete [c, d]"))
				})

				It("fails all", func() {
					ExpectError(constraintset.Match(ctx.WithSelected("a"))).To(MatchError("group \"A\" must be complete [a, b] AND group \"B\" must be complete [c, d]"))
				})
			})

			Context("activated", func() {
				It("or", func() {
					constraintset := constraints.New().Add(constraints.Or(constraints.Activated("c"), constraints.Activated("a")))
					Expect(constraintset.Match(ctx.WithSelected("a", "d"))).To(Equal(constraints.Yes))
				})

				It("and", func() {
					constraintset := constraints.New().Add(constraints.And(constraints.Activated("c"), constraints.Activated("a")))
					Expect(constraintset.Match(ctx.WithSelected("c", "a"))).To(Equal(constraints.Yes))
				})

				It("not", func() {
					constraintset := constraints.New().Add(constraints.Not(constraints.Activated("c")))
					Expect(constraintset.Match(ctx.WithSelected("b", "a"))).To(Equal(constraints.Yes))
				})
			})

			Context("conditional", func() {
				It("succeeds by activation", func() {
					constraintset := constraints.New().Add(constraints.And(constraints.Activated("a"), constraints.Complete("B")))
					Expect(constraintset.Match(ctx.WithSelected("c", "d", "a"))).To(Equal(constraints.Yes))
				})

				It("succeeds by no activation", func() {
					constraintset := constraints.New().Add(constraints.And(constraints.Activated("a"), constraints.Complete("B")))
					Expect(constraintset.Match(ctx.WithSelected("d"))).To(Equal(constraints.No))
				})
			})
		})
	})
})

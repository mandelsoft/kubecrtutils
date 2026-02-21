package types_test

import (
	"github.com/mandelsoft/kubecrtutils/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var MA = types.Mappings{
	"inner": "outer",
	"in":    "out",
}

var _ = Describe("Mapping Test Environment", func() {
	Context("name mappings", func() {
		It("simple", func() {
			Expect(MA.Map("other")).To(Equal("other"))
			Expect(MA.Map("inner")).To(Equal("outer"))
		})

		It("composition", func() {

			orig := types.Mappings{
				"outer": "cli",
			}

			eff := MA.ApplyTo(orig)
			Expect(eff.Map("other")).To(Equal("other"))
			Expect(eff.Map("in")).To(Equal("out"))
			Expect(eff.Map("inner")).To(Equal("cli"))
		})
	})
})

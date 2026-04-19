package controller_test

import (
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/kubecrtutils/controller"
	"github.com/mandelsoft/kubecrtutils/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type setting struct {
	groups map[string][]string
	names  []string
}

var _ types.ControllerSet = (*setting)(nil)

func (s *setting) GetNames() []string {
	return s.names
}

func (s *setting) GetGroups() map[string][]string {
	return s.groups
}

var _ = Describe("Test Environment", func() {

	settings := &setting{
		groups: map[string][]string{
			"A": []string{"B", "C"},
			"B": []string{"a", "b"},
			"C": []string{"c"},
		},
		names: []string{
			"a", "b", "c", "d",
		},
	}
	It("", func() {
		refs := controller.CompleteSet(settings)
		Expect(refs).To(Equal(map[string]set.Set[string]{
			"a": set.New("a"),
			"b": set.New("b"),
			"c": set.New("c"),
			"d": set.New("d"),
			"A": set.New("a", "b", "c"),
			"B": set.New("a", "b"),
			"C": set.New("c"),
		}))
	})
})

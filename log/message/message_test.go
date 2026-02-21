package message

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type LogCatcher struct {
	message string
	args    []interface{}
}

func NewLogCatcher(msg string, args ...interface{}) *LogCatcher {
	return &LogCatcher{message: msg, args: args}
}

func (c *LogCatcher) Enabled(lvl int) bool {
	return true
}

func (c *LogCatcher) Info(msg string, args ...interface{}) {
	c.message = msg
	c.args = args
}

type Element struct {
	base string
	typ  string
	name string
}

func (e *Element) NormalizeTo(i *[]interface{}) {
	*i = append(*i, "typ"+e.base, e.typ, "name"+e.base, e.name)
}

func (e *Element) Message() string {
	return fmt.Sprintf("{{typ%s}} {{name%s}}", e.base, e.base)
}

var _ = Describe("Test Environment", func() {
	var logger *LogCatcher

	BeforeEach(func() {
		logger = &LogCatcher{}
	})

	Context("", func() {
		It("plain", func() {
			Info(logger, "this is a test")
			Expect(logger).To(Equal(NewLogCatcher("this is a test")))
		})

		It("composed plain", func() {
			Info(logger, "this", " is a ", "test")
			Expect(logger).To(Equal(NewLogCatcher("this is a test")))
		})

		It("provider", func() {
			Info(logger, &Element{typ: "cluster", name: "mine"})
			Expect(logger).To(Equal(NewLogCatcher("{{typ}} {{name}}", "typ", "cluster", "name", "mine")))
		})

		It("values", func() {
			Info(logger, "this is a test", Values("name", "mine"))
			Expect(logger).To(Equal(NewLogCatcher("this is a test", "name", "mine")))
		})

		It("mixed", func() {
			Info(logger, &Element{typ: "cluster", name: "mine"}, " is healthy")
			Expect(logger).To(Equal(NewLogCatcher("{{typ}} {{name}} is healthy", "typ", "cluster", "name", "mine")))
		})

		It("value provider", func() {
			Info(logger, "cluster ", KeyValue("cluster", "mine"), " is healthy")
			Expect(logger).To(Equal(NewLogCatcher("cluster {{cluster}} is healthy", "cluster", "mine")))
		})
	})
})

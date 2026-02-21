package kcp

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Path Test Environment", func() {
	Context("path substitution", func() {
		It("handle suffix path", func() {
			Expect(urlPath("/api/clusters", "test")).To(Equal("/api/clusters/test"))
			Expect(urlPath("api/clusters", "test")).To(Equal("api/clusters/test"))
			Expect(urlPath("/api/clusters/", "test")).To(Equal("/api/clusters/test"))
			Expect(urlPath("api/clusters/", "test")).To(Equal("api/clusters/test"))
		})
		It("handle suffix substitution", func() {
			Expect(urlPath("/api/clusters/base", "test")).To(Equal("/api/clusters/test"))
			Expect(urlPath("api/clusters/base", "test")).To(Equal("api/clusters/test"))
			Expect(urlPath("/api/clusters/base/", "test")).To(Equal("/api/clusters/test"))
			Expect(urlPath("api/clusters/base/", "test")).To(Equal("api/clusters/test"))
		})

		It("handle plain path", func() {
			Expect(urlPath("/clusters", "test")).To(Equal("/clusters/test"))
			Expect(urlPath("clusters", "test")).To(Equal("clusters/test"))
			Expect(urlPath("/clusters/", "test")).To(Equal("/clusters/test"))
			Expect(urlPath("clusters/", "test")).To(Equal("clusters/test"))
		})
		It("handle plain substitution", func() {
			Expect(urlPath("/clusters/base", "test")).To(Equal("/clusters/test"))
			Expect(urlPath("clusters/base", "test")).To(Equal("clusters/test"))
			Expect(urlPath("/clusters/base/", "test")).To(Equal("/clusters/test"))
			Expect(urlPath("clusters/base/", "test")).To(Equal("clusters/test"))
		})

		It("fallback", func() {
			Expect(urlPath("/base", "test")).To(Equal("/base"))
			Expect(urlPath("/base/", "test")).To(Equal("/base/"))
			Expect(urlPath("base/", "test")).To(Equal("base/"))
			Expect(urlPath("base", "test")).To(Equal("base"))
			Expect(urlPath("/", "test")).To(Equal("/"))
			Expect(urlPath("", "test")).To(Equal(""))
		})
	})
})

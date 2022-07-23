package notificationservices

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetProvider", func() {
	var (
		provider      string
		expectedValue string
	)
	Context("Provider is slack", func() {
		BeforeEach(func() {
			provider = "slack"
			expectedValue = "slack"
		})
		It("should get a pointer of SlackNotificationServiceProvider that has ProviderName field value = slack", func() {
			nspp := GetProvider(provider)
			Expect(nspp).NotTo(BeNil())
			Expect((*nspp).GetName()).To(Equal(expectedValue))
		})
	})

	Context("Provider is non-supported one", func() {
		BeforeEach(func() {
			provider = "foo"
		})
		It("should get nil pointer ", func() {
			nspp := GetProvider(provider)
			// If it should be nil, use BeNil() instead of Equal(nil) ref: https://onsi.github.io/gomega/
			Expect(nspp).To(BeNil())
		})
	})
})

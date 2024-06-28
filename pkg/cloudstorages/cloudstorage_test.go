package cloudstorages

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetProvider", func() {
	var (
		provider      string
		expectedValue string
	)
	Context("Provider is aws", func() {
		BeforeEach(func() {
			provider = "aws"
			expectedValue = "aws"
		})
		It("should get a pointer of S3CloudStorageProvider that has ProviderName field value = aws", func() {
			cspp := GetProvider(provider)
			Expect(cspp).NotTo(BeNil())
			Expect((*cspp).GetName()).To(Equal(expectedValue))
		})
	})

	Context("Provider is gcp", func() {
		BeforeEach(func() {
			provider = "gcp"
			expectedValue = "gcp"
		})
		It("should get a pointer of GCPCloudStorageProvider that has ProviderName field value = gcp", func() {
			cspp := GetProvider(provider)
			Expect(cspp).NotTo(BeNil())
			Expect((*cspp).GetName()).To(Equal(expectedValue))
		})
	})

	Context("Provider is s3", func() {
		BeforeEach(func() {
			provider = "s3"
			expectedValue = "s3"
		})
		It("should get a pointer of S3CloudStorageProvider that has ProviderName field value = s3", func() {
			cspp := GetProvider(provider)
			Expect(cspp).NotTo(BeNil())
			Expect((*cspp).GetName()).To(Equal(expectedValue))
		})
	})

	Context("Provider is non-supported one", func() {
		BeforeEach(func() {
			provider = "foo"
		})
		It("should get nil pointer ", func() {
			cspp := GetProvider(provider)
			// If it should be nil, use BeNil() instead of Equal(nil) ref: https://onsi.github.io/gomega/
			Expect(cspp).To(BeNil())
		})
	})
})

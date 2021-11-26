package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("getCloudStoragePath", func() {
	var (
		provider    string
		bucket      string
		gatlingName string
		subDir      string
	)

	BeforeEach(func() {
		bucket = "testBucket"
		gatlingName = "testGatling"
		subDir = "subDir"
	})

	Context("Provider values not assumed", func() {
		BeforeEach(func() {
			provider = ""
		})
		It("path is empty", func() {
			Expect(getCloudStoragePath(provider, bucket, gatlingName, subDir)).To(Equal(""))
		})
	})

	Context("Provider value aws", func() {
		BeforeEach(func() {
			provider = "aws"
		})
		It("path is aws s3 bucket", func() {
			Expect(getCloudStoragePath(provider, bucket, gatlingName, subDir)).To(Equal("s3:testBucket/testGatling/subDir"))
		})
	})

	Context("Provider value gcp", func() {
		BeforeEach(func() {
			provider = "gcp"
		})
		It("path is empty", func() {
			Expect(getCloudStoragePath(provider, bucket, gatlingName, subDir)).To(Equal(""))
		})
	})

	Context("Provider value azure", func() {
		BeforeEach(func() {
			provider = "azure"
		})
		It("path is empty", func() {
			Expect(getCloudStoragePath(provider, bucket, gatlingName, subDir)).To(Equal(""))
		})
	})
})

var _ = Describe("getCloudStorageReportURL", func() {
	var (
		provider    string
		bucket      string
		gatlingName string
		subDir      string
	)

	BeforeEach(func() {
		bucket = "testBucket"
		gatlingName = "testGatling"
		subDir = "subDir"
	})

	Context("Provider values not assumed", func() {
		BeforeEach(func() {
			provider = ""
		})
		It("path is empty", func() {
			Expect(getCloudStorageReportURL(provider, bucket, gatlingName, subDir)).To(Equal(""))
		})
	})

	Context("Provider value aws", func() {
		BeforeEach(func() {
			provider = "aws"
		})
		It("path is aws s3 bucket", func() {
			Expect(getCloudStorageReportURL(provider, bucket, gatlingName, subDir)).To(Equal("https://testBucket.s3.amazonaws.com/testGatling/subDir/index.html"))
		})
	})

	Context("Provider value gcp", func() {
		BeforeEach(func() {
			provider = "gcp"
		})
		It("path is empty", func() {
			Expect(getCloudStorageReportURL(provider, bucket, gatlingName, subDir)).To(Equal(""))
		})
	})

	Context("Provider value azure", func() {
		BeforeEach(func() {
			provider = "azure"
		})
		It("path is empty", func() {
			Expect(getCloudStorageReportURL(provider, bucket, gatlingName, subDir)).To(Equal(""))
		})
	})
})

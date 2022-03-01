package cloudstorages

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetName", func() {
	var (
		provider      string
		expectedValue string
	)
	BeforeEach(func() {
		provider = "gcp"
		expectedValue = "gcp"
	})
	Context("Provider is gcp", func() {
		It("should get provider name = gcp", func() {
			csp := &GCPCloudStorageProvider{providerName: provider}
			Expect(csp.GetName()).To(Equal(expectedValue))
		})
	})
})

var _ = Describe("GetCloudStoragePath", func() {
	var (
		provider      string
		bucket        string
		gatlingName   string
		subDir        string
		expectedValue string
	)
	BeforeEach(func() {
		provider = "gcp"
		bucket = "testBucket"
		gatlingName = "testGatling"
		subDir = "subDir"
		expectedValue = "gs://testBucket/testGatling/subDir"
	})
	Context("Provider is gcp", func() {
		It("path is gcp gcs bucket", func() {
			csp := &GCPCloudStorageProvider{providerName: provider}
			Expect(csp.GetCloudStoragePath(bucket, gatlingName, subDir)).To(Equal(expectedValue))
		})
	})
})

var _ = Describe("GetCloudStorageReportURL", func() {
	var (
		provider      string
		bucket        string
		gatlingName   string
		subDir        string
		expectedValue string
	)
	BeforeEach(func() {
		provider = "gcp"
		bucket = "testBucket"
		gatlingName = "testGatling"
		subDir = "subDir"
		expectedValue = "https://storage.googleapis.com/testBucket/testGatling/subDir/index.html"
	})
	Context("Provider is gcp", func() {
		It("path is gcp gcs bucket", func() {
			csp := &GCPCloudStorageProvider{providerName: provider}
			Expect(csp.GetCloudStorageReportURL(bucket, gatlingName, subDir)).To(Equal(expectedValue))
		})
	})
})

var _ = Describe("GetGatlingTransferResultCommand", func() {
	var (
		provider             string
		resultsDirectoryPath string
		region               string
		storagePath          string
		expectedValue        string
	)
	BeforeEach(func() {
		provider = "gcp"
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = ""
		storagePath = "testStoragePath"
		expectedValue = `
RESULTS_DIR_PATH=testResultsDirectoryPath
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
# assumes each pod only contain single gatling log file but use for loop to use find command result
for source in $(find ${RESULTS_DIR_PATH} -type f -name *.log)
do
	rclone copyto ${source} testStoragePath/${HOSTNAME}.log
done
`
	})
	Context("Provider is gcp", func() {
		It("provider is gcp", func() {
			csp := &GCPCloudStorageProvider{providerName: provider}
			Expect(csp.GetGatlingTransferResultCommand(resultsDirectoryPath, region, storagePath)).To(Equal(expectedValue))
		})
	})
})

var _ = Describe("GetGatlingAggregateResultCommand", func() {
	var (
		provider             string
		resultsDirectoryPath string
		region               string
		storagePath          string
		expectedValue        string
	)
	BeforeEach(func() {
		provider = "gcp"
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = ""
		storagePath = "testStoragePath"
		expectedValue = `
GATLING_AGGREGATE_DIR=testResultsDirectoryPath
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
rclone copy testStoragePath ${GATLING_AGGREGATE_DIR}
`
	})
	Context("Provider is gcp", func() {
		It("provider is gcp", func() {
			gcp := &GCPCloudStorageProvider{providerName: provider}
			Expect(gcp.GetGatlingAggregateResultCommand(resultsDirectoryPath, region, storagePath)).To(Equal(expectedValue))
		})
	})
})

var _ = Describe("GetGatlingTransferReportCommand", func() {
	var (
		provider             string
		resultsDirectoryPath string
		region               string
		storagePath          string
		expectedValue        string
	)
	BeforeEach(func() {
		provider = "gcp"
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = ""
		storagePath = "testStoragePath"
		expectedValue = `
GATLING_AGGREGATE_DIR=testResultsDirectoryPath
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" testStoragePath
`
	})
	Context("Provider is gcp", func() {
		It("provider is gcp", func() {
			gcp := &GCPCloudStorageProvider{providerName: provider}
			Expect(gcp.GetGatlingTransferReportCommand(resultsDirectoryPath, region, storagePath)).To(Equal(expectedValue))
		})
	})
})

package cloudstorages

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var storageAccount string = "testaccount"

var cspArgs []EnvVars = []EnvVars{
	{
		{
			Name:  "AZUREBLOB_ACCOUNT",
			Value: storageAccount,
		},
	},
}

var _ = Describe("GetName", func() {
	var (
		provider      string
		expectedValue string
	)
	BeforeEach(func() {
		provider = "azure"
		expectedValue = "azure"
	})
	Context("Provider is azure", func() {
		It("should get provider name = azure", func() {
			csp := &AzureCloudStorageProvider{providerName: provider}
			csp.init(cspArgs)
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
		provider = "azure"
		bucket = "testBucket"
		gatlingName = "testGatling"
		subDir = "subDir"
		expectedValue = fmt.Sprintf("az:%s/%s/%s", bucket, gatlingName, subDir)
	})
	Context("Provider is azure", func() {
		It("path is azure blob storage container", func() {
			csp := &AzureCloudStorageProvider{providerName: provider}
			csp.init(cspArgs)
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
		provider = "azure"
		bucket = "testBucket"
		gatlingName = "testGatling"
		subDir = "subDir"
		expectedValue = fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s/%s/index.html",
			storageAccount, bucket, gatlingName, subDir)
	})
	Context("Provider is azure", func() {
		It("path is azure s3 bucket", func() {
			csp := &AzureCloudStorageProvider{providerName: provider}
			csp.init(cspArgs)
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
		provider = "azure"
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = ""
		storagePath = "testStoragePath"
		expectedValue = fmt.Sprintf(`
export RCLONE_AZUREBLOB_ACCOUNT=${AZUREBLOB_ACCOUNT}
export RCLONE_AZUREBLOB_KEY=${AZUREBLOB_KEY}
export RCLONE_AZUREBLOB_SAS_URL=${AZUREBLOB_SAS_URL}
RESULTS_DIR_PATH=%s
rclone config create az azureblob env_auth=true
while true; do
  if [ -f "${RESULTS_DIR_PATH}/FAILED" ]; then
    echo "Skip transfering gatling results"
    break
  fi
  if [ -f "${RESULTS_DIR_PATH}/COMPLETED" ]; then
    for source in $(find ${RESULTS_DIR_PATH} -type f -name *.log)
    do
      rclone copyto ${source} %s/${HOSTNAME}.log
    done
    break
  fi
  sleep 1;
done	
`, resultsDirectoryPath, storagePath)
	})
	Context("Provider is azure", func() {
		It("returns commands with azure blob storage rclone config", func() {
			csp := &AzureCloudStorageProvider{providerName: provider}
			csp.init(cspArgs)
			fmt.Println(csp.GetGatlingTransferResultCommand(resultsDirectoryPath, region, storagePath))
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
		provider = "azure"
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = ""
		storagePath = "testStoragePath"
		expectedValue = fmt.Sprintf(`
export RCLONE_AZUREBLOB_ACCOUNT=${AZUREBLOB_ACCOUNT}
export RCLONE_AZUREBLOB_KEY=${AZUREBLOB_KEY}
export RCLONE_AZUREBLOB_SAS_URL=${AZUREBLOB_SAS_URL}
GATLING_AGGREGATE_DIR=%s
rclone config create az azureblob env_auth=true
rclone copy %s ${GATLING_AGGREGATE_DIR}
`, resultsDirectoryPath, storagePath)
	})
	Context("Provider is azure", func() {
		It("returns commands with azure blob storage rclone config", func() {
			csp := &AzureCloudStorageProvider{providerName: provider}
			csp.init(cspArgs)
			Expect(csp.GetGatlingAggregateResultCommand(resultsDirectoryPath, region, storagePath)).To(Equal(expectedValue))
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
		provider = "azure"
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = ""
		storagePath = "testStoragePath"
		expectedValue = fmt.Sprintf(`
export RCLONE_AZUREBLOB_ACCOUNT=${AZUREBLOB_ACCOUNT}
export RCLONE_AZUREBLOB_KEY=${AZUREBLOB_KEY}
export RCLONE_AZUREBLOB_SAS_URL=${AZUREBLOB_SAS_URL}
GATLING_AGGREGATE_DIR=%s
rclone config create az azureblob env_auth=true
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" %s
`, resultsDirectoryPath, storagePath)
	})
	Context("Provider is azure", func() {
		It("returns commands with azure blob storage rclone config", func() {
			csp := &AzureCloudStorageProvider{providerName: provider}
			Expect(csp.GetGatlingTransferReportCommand(resultsDirectoryPath, region, storagePath)).To(Equal(expectedValue))
		})
	})
})

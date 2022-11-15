package cloudstorages

import (
	"fmt"
)

type AzureCloudStorageProvider struct {
	providerName   string
	storageAccount string
}

func (p *AzureCloudStorageProvider) init(args []EnvVars) {
	if len(args) > 0 {
		var envs EnvVars = args[0]
		for _, env := range envs {
			if env.Name == "AZUREBLOB_ACCOUNT" {
				p.storageAccount = env.Value
				break
			}
		}
	}
}

func (p *AzureCloudStorageProvider) GetName() string {
	return p.providerName
}

func (p *AzureCloudStorageProvider) GetCloudStoragePath(bucket string, gatlingName string, subDir string) string {
	// Format azureblob:<bucket>/<gatling-name>/<sub-dir>
	return fmt.Sprintf("az:%s/%s/%s", bucket, gatlingName, subDir)
}

func (p *AzureCloudStorageProvider) GetCloudStorageReportURL(bucket string, gatlingName string, subDir string) string {
	// Format https://<storage-account>.blob.core.windows.net/bucket/<gatling-name>/subdir/hoge.log
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s/%s/index.html", p.storageAccount, bucket, gatlingName, subDir)
}

func (p *AzureCloudStorageProvider) GetGatlingTransferResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	// region param isn't used
	template := `
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
`
	return fmt.Sprintf(template, resultsDirectoryPath, storagePath)
}

func (p *AzureCloudStorageProvider) GetGatlingAggregateResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	// region param isn't used
	template := `
export RCLONE_AZUREBLOB_ACCOUNT=${AZUREBLOB_ACCOUNT}
export RCLONE_AZUREBLOB_KEY=${AZUREBLOB_KEY}
export RCLONE_AZUREBLOB_SAS_URL=${AZUREBLOB_SAS_URL}
GATLING_AGGREGATE_DIR=%s
rclone config create az azureblob env_auth=true
rclone copy %s ${GATLING_AGGREGATE_DIR}
`
	return fmt.Sprintf(template, resultsDirectoryPath, storagePath)
}

func (p *AzureCloudStorageProvider) GetGatlingTransferReportCommand(resultsDirectoryPath string, region string, storagePath string) string {
	// region param isn't used
	template := `
export RCLONE_AZUREBLOB_ACCOUNT=${AZUREBLOB_ACCOUNT}
export RCLONE_AZUREBLOB_KEY=${AZUREBLOB_KEY}
export RCLONE_AZUREBLOB_SAS_URL=${AZUREBLOB_SAS_URL}
GATLING_AGGREGATE_DIR=%s
rclone config create az azureblob env_auth=true
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" %s
`
	return fmt.Sprintf(template, resultsDirectoryPath, storagePath)
}

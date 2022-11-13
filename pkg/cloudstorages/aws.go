package cloudstorages

import (
	"fmt"
)

type AWSCloudStorageProvider struct {
	providerName string
}

func (p *AWSCloudStorageProvider) init(args []EnvVars) { /* do nothing */ }

func (p *AWSCloudStorageProvider) GetName() string {
	return p.providerName
}

func (p *AWSCloudStorageProvider) GetCloudStoragePath(bucket string, gatlingName string, subDir string) string {
	// Format s3:<bucket>/<gatling-name>/<sub-dir>
	return fmt.Sprintf("s3:%s/%s/%s", bucket, gatlingName, subDir)
}

func (p *AWSCloudStorageProvider) GetCloudStorageReportURL(bucket string, gatlingName string, subDir string) string {
	// Format https://<bucket>.s3.amazonaws.com/<gatling-name>/<sub-dir>/index.html
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s/%s/index.html", bucket, gatlingName, subDir)
}

func (p *AWSCloudStorageProvider) GetGatlingTransferResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
RESULTS_DIR_PATH=%s
rclone config create s3 s3 env_auth=true region %s
while true; do
  if [ -f "${RESULTS_DIR_PATH}/COMPLETED" ]; then
    for source in $(find ${RESULTS_DIR_PATH} -type f -name *.log)
    do
      rclone copyto ${source} --s3-no-check-bucket --s3-env-auth %s/${HOSTNAME}.log
    done
    break
  fi
  sleep 1;
done
`
	return fmt.Sprintf(template, resultsDirectoryPath, region, storagePath)
}

func (p *AWSCloudStorageProvider) GetGatlingAggregateResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
rclone config create s3 s3 env_auth=true region %s
rclone copy --s3-no-check-bucket --s3-env-auth %s ${GATLING_AGGREGATE_DIR}
`
	return fmt.Sprintf(template, resultsDirectoryPath, region, storagePath)
}

func (p *AWSCloudStorageProvider) GetGatlingTransferReportCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
rclone config create s3 s3 env_auth=true region %s
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" --s3-no-check-bucket --s3-env-auth %s
`
	return fmt.Sprintf(template, resultsDirectoryPath, region, storagePath)
}

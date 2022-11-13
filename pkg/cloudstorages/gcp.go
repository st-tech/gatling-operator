package cloudstorages

import (
	"fmt"
)

type GCPCloudStorageProvider struct {
	providerName string
}

func (p *GCPCloudStorageProvider) init(args []EnvVars) { /* do nothing */ }

func (p *GCPCloudStorageProvider) GetName() string {
	return p.providerName
}

func (p *GCPCloudStorageProvider) GetCloudStoragePath(bucket string, gatlingName string, subDir string) string {
	return fmt.Sprintf("gs://%s/%s/%s", bucket, gatlingName, subDir)
}

func (p *GCPCloudStorageProvider) GetCloudStorageReportURL(bucket string, gatlingName string, subDir string) string {
	// Format http(s)://<bucket>.storage.googleapis.com/<gatling-name>/<sub-dir>/index.html
	// or http(s)://storage.googleapis.com/<bucket>/<gatling-name>/<sub-dir>/index.html
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s/%s/index.html", bucket, gatlingName, subDir)
}

// param: region is not used in GCP GCS, thus set just dummy value
func (p *GCPCloudStorageProvider) GetGatlingTransferResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
RESULTS_DIR_PATH=%s
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
while true; do
  if [ -f "${RESULTS_DIR_PATH}/COMPLETED" ]; then
    # assumes each pod only contain single gatling log file but use for loop to use find command result
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

// param: region is not used in GCP GCS, thus set just dummy value
func (p *GCPCloudStorageProvider) GetGatlingAggregateResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
rclone copy %s ${GATLING_AGGREGATE_DIR}
`
	return fmt.Sprintf(template, resultsDirectoryPath, storagePath)
}

// param: region is not used in GCP GCS, thus set just dummy value
func (p *GCPCloudStorageProvider) GetGatlingTransferReportCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" %s
`
	return fmt.Sprintf(template, resultsDirectoryPath, storagePath)
}

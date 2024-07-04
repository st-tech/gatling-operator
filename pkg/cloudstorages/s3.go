package cloudstorages

import (
	"fmt"
	"strings"
)

type S3CloudStorageProvider struct {
	providerName         string
	customS3ProviderHost string
}

func (p *S3CloudStorageProvider) init(args []EnvVars) {
	if len(args) > 0 {
		var envs EnvVars = args[0]
		for _, env := range envs {
			if env.Name == "RCLONE_S3_ENDPOINT" {
				p.customS3ProviderHost = p.checkAndRemoveProtocol(env.Value)
				break
			}
		}
	}
}

func (p *S3CloudStorageProvider) checkAndRemoveProtocol(url string) string {
	idx := strings.Index(url, "://")
	if idx == -1 {
		return url
	}
	return url[idx+3:]
}

func (p *S3CloudStorageProvider) GetName() string {
	return p.providerName
}

func (p *S3CloudStorageProvider) GetCloudStoragePath(bucket string, gatlingName string, subDir string) string {
	// Format s3:<bucket>/<gatling-name>/<sub-dir>
	return fmt.Sprintf("s3:%s/%s/%s", bucket, gatlingName, subDir)
}

func (p *S3CloudStorageProvider) GetCloudStorageReportURL(bucket string, gatlingName string, subDir string) string {
	// Format https://<bucket>.<s3-provider-host>/<gatling-name>/<sub-dir>/index.html
	defaultS3ProviderHost := "s3.amazonaws.com"
	s3ProviderHost := defaultS3ProviderHost
	if p.customS3ProviderHost != "" {
		s3ProviderHost = p.customS3ProviderHost
	}

	return fmt.Sprintf("https://%s.%s/%s/%s/index.html", bucket, s3ProviderHost, gatlingName, subDir)
}

func (p *S3CloudStorageProvider) GetGatlingTransferResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
RESULTS_DIR_PATH=%s
rclone config create s3 s3 env_auth=true region %s
while true; do
  if [ -f "${RESULTS_DIR_PATH}/FAILED" ]; then
    echo "Skip transfering gatling results"
    break
  fi
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

func (p *S3CloudStorageProvider) GetGatlingAggregateResultCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
rclone config create s3 s3 env_auth=true region %s
rclone copy --s3-no-check-bucket --s3-env-auth %s ${GATLING_AGGREGATE_DIR}
`
	return fmt.Sprintf(template, resultsDirectoryPath, region, storagePath)
}

func (p *S3CloudStorageProvider) GetGatlingTransferReportCommand(resultsDirectoryPath string, region string, storagePath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
rclone config create s3 s3 env_auth=true region %s
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" --s3-no-check-bucket --s3-env-auth %s
`
	return fmt.Sprintf(template, resultsDirectoryPath, region, storagePath)
}

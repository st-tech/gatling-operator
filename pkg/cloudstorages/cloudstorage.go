package cloudstorages

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
)

type EnvVars []corev1.EnvVar

type CloudStorageProvider interface {
	init(args []EnvVars)
	GetName() string
	GetCloudStoragePath(bucket string, gatlingName string, subDir string) string
	GetCloudStorageReportURL(bucket string, gatlingName string, subDir string) string
	GetGatlingTransferResultCommand(resultsDirectoryPath string, region string, storagePath string) string
	GetGatlingAggregateResultCommand(resultsDirectoryPath string, region string, storagePath string) string
	GetGatlingTransferReportCommand(resultsDirectoryPath string, region string, storagePath string) string
}

// use sync.Map to achieve thread safe read and write to map
var cloudStorageProvidersSyncMap = &sync.Map{}

func GetProvider(provider string, args ...EnvVars) *CloudStorageProvider {
	v, ok := cloudStorageProvidersSyncMap.Load(provider)
	if !ok {
		var csp CloudStorageProvider
		switch provider {
		case "aws":
			csp = &AWSCloudStorageProvider{providerName: provider}
		case "gcp":
			csp = &GCPCloudStorageProvider{providerName: provider}
		case "azure":
			csp = &AzureCloudStorageProvider{providerName: provider}
		default:
			return nil
		}
		csp.init(args)
		v, _ = cloudStorageProvidersSyncMap.LoadOrStore(provider, &csp)
	}
	return v.(*CloudStorageProvider)
}

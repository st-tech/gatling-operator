package controllers

import (
	"fmt"
)

func getCloudStoragePath(provider string, bucket string, gatlingName string, subDir string) string {
	switch provider {
	case "aws":
		// Format s3:<bucket>/<gatling-name>/<sub-dir>
		return fmt.Sprintf("s3:%s/%s/%s", bucket, gatlingName, subDir)
	case "gcp": //not supported yet
		return ""
	case "azure": //not supported yet
		return ""
	default:
		return ""
	}
}

func getCloudStorageReportURL(provider string, bucket string, gatlingName string, subDir string) string {
	switch provider {
	case "aws":
		// Format https://<bucket>.s3.amazonaws.com/<gatling-name>/<sub-dir>/index.html
		return fmt.Sprintf("https://%s.s3.amazonaws.com/%s/%s/index.html", bucket, gatlingName, subDir)
	case "gcp": //not supported yet
		// Format http(s)://<bucket>.storage.googleapis.com/<gatling-name>/<sub-dir>/index.html
		// or http(s)://storage.googleapis.com/<bucket>/<gatling-name>/<sub-dir>/index.html
		return ""
	case "azure": //not supported yet
		// Format https://<bucket>.blob.core.windows.net/<gatling-name>/<sub-dir>/index.html
		return ""
	default:
		return ""
	}
}

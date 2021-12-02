package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("getGatlingGenerateReportCommand", func() {
	var (
		resultDirectoryPath string
		template            string
	)

	BeforeEach(func() {
		resultDirectoryPath = "testDirectoryPath"
		template = `
GATLING_AGGREGATE_DIR=testDirectoryPath
DIR_NAME=$(dirname ${GATLING_AGGREGATE_DIR})
BASE_NAME=$(basename ${GATLING_AGGREGATE_DIR})
gatling.sh -rf ${DIR_NAME} -ro ${BASE_NAME}
`
	})

	It("getExceptValue", func() {
		Expect(getGatlingGenerateReportCommand(resultDirectoryPath)).NotTo(Equal(template))
	})
})

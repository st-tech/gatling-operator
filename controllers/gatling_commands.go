package controllers

import (
	"fmt"
)

func getGatlingRunnerCommand(
	simulationsDirectoryPath string, tempSimulationsDirectoryPath string, resourcesDirectoryPath string,
	resultsDirectoryPath string, startTime string, simulationClass string) string {

	template := `
SIMULATIONS_DIR_PATH="%s"
TEMP_SIMULATIONS_DIR_PATH="%s"
RESOURCES_DIR_PATH="%s"
RESULTS_DIR_PATH="%s"
START_TIME="%s"
if [ -z "${START_TIME}" ];
then
  START_TIME=$(date +"%%Y-%%m-%%d %%H:%%M:%%S" --utc)
fi
start_time_stamp=$(date -d "${START_TIME}" +"%%s")
current_time_stamp=$(date +"%%s")
echo "wait until ${START_TIME}"
until [ ${current_time_stamp} -ge ${start_time_stamp} ];
do
  current_time_stamp=$(date +"%%s")
  echo "it's ${current_time_stamp} now and waiting until ${start_time_stamp} ..."
  sleep 1;
done
if [ ! -d ${SIMULATIONS_DIR_PATH} ];
then
  mkdir -p ${SIMULATIONS_DIR_PATH}
fi
if [ -d ${TEMP_SIMULATIONS_DIR_PATH} ];
then
	cp -p ${TEMP_SIMULATIONS_DIR_PATH}/*.scala ${SIMULATIONS_DIR_PATH}
fi
if [ ! -d ${RESOURCES_DIR_PATH} ];
then
  mkdir -p ${RESOURCES_DIR_PATH}
fi
if [ ! -d ${RESULTS_DIR_PATH} ];
then
  mkdir -p ${RESULTS_DIR_PATH}
fi
gatling.sh -sf ${SIMULATIONS_DIR_PATH} -s %s -rsf ${RESOURCES_DIR_PATH} -rf ${RESULTS_DIR_PATH} -nr
`
	return fmt.Sprintf(template,
		simulationsDirectoryPath,
		tempSimulationsDirectoryPath,
		resourcesDirectoryPath,
		resultsDirectoryPath,
		startTime,
		simulationClass)
}

func getGatlingTransferResultCommand(resultsDirectoryPath string, reportStorage string) string {
	template := `
RESULTS_DIR_PATH="%s"
rclone config create s3 s3 env_auth=true
for source in $(find ${RESULTS_DIR_PATH} -type f -name *.log)
do
	rclone copyto ${source} --s3-env-auth %s/${HOSTNAME}.log
done
`
	return fmt.Sprintf(template, resultsDirectoryPath, reportStorage)
}

func getGatlingAggregateResultCommand(resultsDirectoryPath string, reportStorage string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
rclone config create s3 s3 env_auth=true
rclone copy --s3-env-auth %s ${GATLING_AGGREGATE_DIR}
`
	return fmt.Sprintf(template, resultsDirectoryPath, reportStorage)
}

func getGatlingGenerateReportCommand(resultsDirectoryPath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
DIR_NAME=$(dirname ${GATLING_AGGREGATE_DIR})
BASE_NAME=$(basename ${GATLING_AGGREGATE_DIR})
gatling.sh -rf ${DIR_NAME} -ro ${BASE_NAME}
`
	return fmt.Sprintf(template, resultsDirectoryPath)
}

func getGatlingTransferReportCommand(resultsDirectoryPath string, reportStorage string) string {
	template := `
rclone config create s3 s3 env_auth=true
GATLING_AGGREGATE_DIR=%s
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" --s3-env-auth %s 
`
	return fmt.Sprintf(template, resultsDirectoryPath, reportStorage)
}

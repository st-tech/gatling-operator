package commands

import (
	"fmt"

	cloudstorages "github.com/st-tech/gatling-operator/pkg/cloudstorages"
)

func GetGatlingWaiterCommand(parallelism *int32, gatlingNamespace string, gatlingName string) string {
	template := `
PARALLELISM=%d
NAMESPACE=%s
JOB_NAME=%s
POD_NAME=$(cat /etc/pod-info/name)

kubectl label pods -n $NAMESPACE $POD_NAME gatling-waiter=initialized

while true; do
  READY_PODS=$(kubectl get pods -n $NAMESPACE --selector=job-name=$JOB_NAME-runner,gatling-waiter=initialized --no-headers | grep -c ".*");
  echo "$READY_PODS/$PARALLELISM pods are ready";
  if  [ $READY_PODS -eq $PARALLELISM ]; then
    break;
  fi;
  sleep 1;
done
`
	return fmt.Sprintf(template,
		*parallelism,
		gatlingNamespace,
		gatlingName,
	)
}

func GetGatlingRunnerCommand(
	simulationsDirectoryPath string, tempSimulationsDirectoryPath string, resourcesDirectoryPath string,
	resultsDirectoryPath string, startTime string, simulationClass string, generateLocalReport bool) string {

	template := `
SIMULATIONS_DIR_PATH=%s
TEMP_SIMULATIONS_DIR_PATH=%s
RESOURCES_DIR_PATH=%s
RESULTS_DIR_PATH=%s
START_TIME="%s"
RUN_STATUS_FILE="${RESULTS_DIR_PATH}/COMPLETED"
if [ -z "${START_TIME}" ]; then
  START_TIME=$(date +"%%Y-%%m-%%d %%H:%%M:%%S" --utc)
fi
start_time_stamp=$(date -d "${START_TIME}" +"%%s")
current_time_stamp=$(date +"%%s")
echo "Wait until ${START_TIME}"
until [ ${current_time_stamp} -ge ${start_time_stamp} ];
do
  current_time_stamp=$(date +"%%s")
  echo "it's ${current_time_stamp} now and waiting until ${start_time_stamp} ..."
  sleep 1;
done
if [ ! -d ${SIMULATIONS_DIR_PATH} ]; then
  mkdir -p ${SIMULATIONS_DIR_PATH}
fi
if [ -d ${TEMP_SIMULATIONS_DIR_PATH} ]; then
  cp -p ${TEMP_SIMULATIONS_DIR_PATH}/*.scala ${SIMULATIONS_DIR_PATH}
fi
if [ ! -d ${RESOURCES_DIR_PATH} ]; then
  mkdir -p ${RESOURCES_DIR_PATH}
fi
if [ ! -d ${RESULTS_DIR_PATH} ]; then
  mkdir -p ${RESULTS_DIR_PATH}
fi
gatling.sh -sf ${SIMULATIONS_DIR_PATH} -s %s -rsf ${RESOURCES_DIR_PATH} -rf ${RESULTS_DIR_PATH} %s %s

GATLING_EXIT_STATUS=$?
if [ $GATLING_EXIT_STATUS -ne 0 ]; then
  RUN_STATUS_FILE="${RESULTS_DIR_PATH}/FAILED"
  echo "gatling.sh has failed!" 1>&2
fi
touch ${RUN_STATUS_FILE}
exit $GATLING_EXIT_STATUS
`
	generateLocalReportOption := "-nr"
	if generateLocalReport {
		generateLocalReportOption = ""
	}

	runModeOptionLocal := "-rm local"

	return fmt.Sprintf(template,
		simulationsDirectoryPath,
		tempSimulationsDirectoryPath,
		resourcesDirectoryPath,
		resultsDirectoryPath,
		startTime,
		simulationClass,
		generateLocalReportOption,
		runModeOptionLocal)
}

func GetGatlingTransferResultCommand(resultsDirectoryPath string, provider string, region string, storagePath string) string {
	var command string
	cspp := cloudstorages.GetProvider(provider)
	if cspp != nil {
		command = (*cspp).GetGatlingTransferResultCommand(resultsDirectoryPath, region, storagePath)
	}
	return command
}

func GetGatlingAggregateResultCommand(resultsDirectoryPath string, provider string, region string, storagePath string) string {
	var command string
	cspp := cloudstorages.GetProvider(provider)
	if cspp != nil {
		command = (*cspp).GetGatlingAggregateResultCommand(resultsDirectoryPath, region, storagePath)
	}
	return command
}

func GetGatlingGenerateReportCommand(resultsDirectoryPath string) string {
	template := `
GATLING_AGGREGATE_DIR=%s
DIR_NAME=$(dirname ${GATLING_AGGREGATE_DIR})
BASE_NAME=$(basename ${GATLING_AGGREGATE_DIR})
gatling.sh -rf ${DIR_NAME} -ro ${BASE_NAME}
`
	return fmt.Sprintf(template, resultsDirectoryPath)
}

func GetGatlingTransferReportCommand(resultsDirectoryPath string, provider string, region string, storagePath string) string {
	var command string
	cspp := cloudstorages.GetProvider(provider)
	if cspp != nil {
		command = (*cspp).GetGatlingTransferReportCommand(resultsDirectoryPath, region, storagePath)
	}
	return command
}

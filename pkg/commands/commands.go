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
	simulationsFormat string, simulationsDirectoryPath string, tempSimulationsDirectoryPath string,
	resourcesDirectoryPath string, resultsDirectoryPath string, startTime string, simulationClass string,
	generateLocalReport bool) string {

	template := `
can_use_gatling_3_11_syntax() {
  version=$1

  # if the version can't be find/parsed, it's better to allow use of newer syntax
  if [ -z "${version}" ]; then
	echo 0
	return
  fi

  ver_major=$(echo "$version" | cut -d. -f1)
  ver_minor=$(echo "$version" | cut -d. -f2)

  # Compare major versions
  if [ "$ver_major" -gt "3" ]; then
	echo 0
	return
  elif [ "$ver_major" -lt "3" ]; then
	echo 1
	return
  fi

  # Compare minor versions
  if [ "$ver_minor" -gt "10" ]; then
	echo 0
	return
  elif [ "$ver_minor" -lt "10" ]; then
	echo 1
	return
  fi

  # you can't use 3.11 syntax, you're running 3.10.x
  echo 1
}

SIMULATIONS_FORMAT=%s
SIMULATIONS_DIR_PATH=%s
TEMP_SIMULATIONS_DIR_PATH=%s
RESOURCES_DIR_PATH=%s
RESULTS_DIR_PATH=%s
START_TIME="%s"
SIMULATION_CLASS=%s
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

if [ ${SIMULATIONS_FORMAT} = "bundle" ]; then
  gatling.sh -sf ${SIMULATIONS_DIR_PATH} -s ${SIMULATION_CLASS} -rsf ${RESOURCES_DIR_PATH} -rf ${RESULTS_DIR_PATH} %s %s
elif [ ${SIMULATIONS_FORMAT} = "gradle" ]; then
  gatling_ver=$(find . -name "build.gradle*" -execdir gradle buildEnvironment \; | grep 'gatling-gradle-plugin:' | sed -n 's/.*:gatling-gradle-plugin:\(.*\)$/\1/p')
  if [ $(can_use_gatling_3_11_syntax "${gatling_ver}") -eq 0 ]; then
    gradle -Dgatling.core.directory.results=${RESULTS_DIR_PATH} gatlingRun --simulation=${SIMULATION_CLASS}
  else
    gradle -Dgatling.core.directory.results=${RESULTS_DIR_PATH} gatlingRun-${SIMULATION_CLASS}
  fi
fi

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
		simulationsFormat,
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

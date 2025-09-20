package commands

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetGatlingWaiterCommand", func() {
	It("getExceptValue", func() {
		expectedValue := `
PARALLELISM=1
NAMESPACE=gatling-system
JOB_NAME=gatling-test
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
		var parallelism int32 = 1
		Expect(GetGatlingWaiterCommand(&parallelism, "gatling-system", "gatling-test")).To(Equal(expectedValue))
	})
})

var _ = Describe("GetGatlingRunnerCommand", func() {
	var (
		simulationsFormat            string
		simulationsDirectoryPath     string
		tempSimulationsDirectoryPath string
		resourcesDirectoryPath       string
		resultsDirectoryPath         string
		startTime                    string
		simulationClass              string
		generateLocalReport          bool
		expectedValue                string
	)

	BeforeEach(func() {
		simulationsFormat = "bundle"
		simulationsDirectoryPath = "testSimulationDirectoryPath"
		tempSimulationsDirectoryPath = "testTempSimulationsDirectoryPath"
		resourcesDirectoryPath = "testResourcesDirectoryPath"
		resultsDirectoryPath = "testResultsDirectoryPath"
		startTime = "2021-09-10 08:45:31"
		simulationClass = "testSimulationClass"
	})

	It("GetCommandsWithLocalReport", func() {
		generateLocalReport = true
		expectedValue = `
SIMULATIONS_FORMAT=bundle
SIMULATIONS_DIR_PATH=testSimulationDirectoryPath
TEMP_SIMULATIONS_DIR_PATH=testTempSimulationsDirectoryPath
RESOURCES_DIR_PATH=testResourcesDirectoryPath
RESULTS_DIR_PATH=testResultsDirectoryPath
START_TIME="2021-09-10 08:45:31"
SIMULATION_CLASS=testSimulationClass
RUN_STATUS_FILE="${RESULTS_DIR_PATH}/COMPLETED"
if [ -z "${START_TIME}" ]; then
  START_TIME=$(date +"%Y-%m-%d %H:%M:%S" --utc)
fi
start_time_stamp=$(date -d "${START_TIME}" +"%s")
current_time_stamp=$(date +"%s")
echo "Wait until ${START_TIME}"
until [ ${current_time_stamp} -ge ${start_time_stamp} ];
do
  current_time_stamp=$(date +"%s")
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
  gatling.sh -sf ${SIMULATIONS_DIR_PATH} -s ${SIMULATION_CLASS} -rsf ${RESOURCES_DIR_PATH} -rf ${RESULTS_DIR_PATH}  -rm local
elif [ ${SIMULATIONS_FORMAT} = "gradle" ]; then
  gradle -Dgatling.core.directory.results=${RESULTS_DIR_PATH} --simulation=${SIMULATION_CLASS}
fi

GATLING_EXIT_STATUS=$?
if [ $GATLING_EXIT_STATUS -ne 0 ]; then
  RUN_STATUS_FILE="${RESULTS_DIR_PATH}/FAILED"
  echo "gatling.sh has failed!" 1>&2
fi
touch ${RUN_STATUS_FILE}
exit $GATLING_EXIT_STATUS
`
		Expect(GetGatlingRunnerCommand(simulationsFormat, simulationsDirectoryPath, tempSimulationsDirectoryPath, resourcesDirectoryPath, resultsDirectoryPath, startTime, simulationClass, generateLocalReport)).To(Equal(expectedValue))
	})

	It("GetCommandWithoutLocalReport", func() {
		generateLocalReport = false
		expectedValue = `
SIMULATIONS_FORMAT=bundle
SIMULATIONS_DIR_PATH=testSimulationDirectoryPath
TEMP_SIMULATIONS_DIR_PATH=testTempSimulationsDirectoryPath
RESOURCES_DIR_PATH=testResourcesDirectoryPath
RESULTS_DIR_PATH=testResultsDirectoryPath
START_TIME="2021-09-10 08:45:31"
SIMULATION_CLASS=testSimulationClass
RUN_STATUS_FILE="${RESULTS_DIR_PATH}/COMPLETED"
if [ -z "${START_TIME}" ]; then
  START_TIME=$(date +"%Y-%m-%d %H:%M:%S" --utc)
fi
start_time_stamp=$(date -d "${START_TIME}" +"%s")
current_time_stamp=$(date +"%s")
echo "Wait until ${START_TIME}"
until [ ${current_time_stamp} -ge ${start_time_stamp} ];
do
  current_time_stamp=$(date +"%s")
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
  gatling.sh -sf ${SIMULATIONS_DIR_PATH} -s ${SIMULATION_CLASS} -rsf ${RESOURCES_DIR_PATH} -rf ${RESULTS_DIR_PATH} -nr -rm local
elif [ ${SIMULATIONS_FORMAT} = "gradle" ]; then
  gradle -Dgatling.core.directory.results=${RESULTS_DIR_PATH} --simulation=${SIMULATION_CLASS}
fi

GATLING_EXIT_STATUS=$?
if [ $GATLING_EXIT_STATUS -ne 0 ]; then
  RUN_STATUS_FILE="${RESULTS_DIR_PATH}/FAILED"
  echo "gatling.sh has failed!" 1>&2
fi
touch ${RUN_STATUS_FILE}
exit $GATLING_EXIT_STATUS
`
		Expect(GetGatlingRunnerCommand(simulationsFormat, simulationsDirectoryPath, tempSimulationsDirectoryPath, resourcesDirectoryPath, resultsDirectoryPath, startTime, simulationClass, generateLocalReport)).To(Equal(expectedValue))
	})
})

var _ = Describe("GetGatlingTransferResultCommand", func() {
	var (
		resultsDirectoryPath string
		provider             string
		region               string
		storagePath          string
		expectedValue        string
	)

	BeforeEach(func() {
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = "ap-northeast-1"
		storagePath = "testStoragePath"
	})

	Context("Provider is aws", func() {
		BeforeEach(func() {
			provider = "aws"
			expectedValue = `
RESULTS_DIR_PATH=testResultsDirectoryPath
rclone config create s3 s3 env_auth=true region ap-northeast-1
while true; do
  if [ -f "${RESULTS_DIR_PATH}/FAILED" ]; then
    echo "Skip transfering gatling results"
    break
  fi
  if [ -f "${RESULTS_DIR_PATH}/COMPLETED" ]; then
    for source in $(find ${RESULTS_DIR_PATH} -type f -name *.log)
    do
      rclone copyto ${source} --s3-no-check-bucket --s3-env-auth testStoragePath/${HOSTNAME}.log
    done
    break
  fi
  sleep 1;
done
`
		})
		It("provider is aws", func() {
			Expect(GetGatlingTransferResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is gcp", func() {
		BeforeEach(func() {
			provider = "gcp"
			expectedValue = `
RESULTS_DIR_PATH=testResultsDirectoryPath
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
while true; do
  if [ -f "${RESULTS_DIR_PATH}/FAILED" ]; then
    echo "Skip transfering gatling results"
    break
  fi
  if [ -f "${RESULTS_DIR_PATH}/COMPLETED" ]; then
    # assumes each pod only contain single gatling log file but use for loop to use find command result
    for source in $(find ${RESULTS_DIR_PATH} -type f -name *.log)
    do
      rclone copyto ${source} testStoragePath/${HOSTNAME}.log
    done
    break
  fi
  sleep 1;
done
`
		})
		It("Provider is gcp", func() {
			Expect(GetGatlingTransferResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is non-supported one", func() {
		BeforeEach(func() {
			provider = "foo"
			expectedValue = ""
		})
		It("Provide is non-supported one", func() {
			Expect(GetGatlingTransferResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is empty", func() {
		BeforeEach(func() {
			provider = ""
			expectedValue = ""
		})
		It("Provider is empty", func() {
			Expect(GetGatlingTransferResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})
})

var _ = Describe("GetGatlingAggregateResultCommand", func() {
	var (
		resultsDirectoryPath string
		provider             string
		region               string
		storagePath          string
		expectedValue        string
	)

	BeforeEach(func() {
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = "ap-northeast-1"
		storagePath = "testStoragePath"
	})

	Context("Provider is aws", func() {
		BeforeEach(func() {
			provider = "aws"
			expectedValue = `
GATLING_AGGREGATE_DIR=testResultsDirectoryPath
rclone config create s3 s3 env_auth=true region ap-northeast-1
rclone copy --s3-no-check-bucket --s3-env-auth testStoragePath ${GATLING_AGGREGATE_DIR}
`
		})
		It("provider is aws", func() {
			Expect(GetGatlingAggregateResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is gcp", func() {
		BeforeEach(func() {
			provider = "gcp"
			expectedValue = `
GATLING_AGGREGATE_DIR=testResultsDirectoryPath
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
rclone copy testStoragePath ${GATLING_AGGREGATE_DIR}
`
		})
		It("Provider is gcp", func() {
			Expect(GetGatlingAggregateResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is non-supported one", func() {
		BeforeEach(func() {
			provider = "foo"
			expectedValue = ""
		})
		It("Provide is non-supported one", func() {
			Expect(GetGatlingAggregateResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is empty", func() {
		BeforeEach(func() {
			provider = ""
			expectedValue = ""
		})
		It("Provider is empty", func() {
			Expect(GetGatlingAggregateResultCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})
})

var _ = Describe("GetGatlingGenerateReportCommand", func() {
	var (
		resultsDirectoryPath string
		expectedValue        string
	)

	BeforeEach(func() {
		resultsDirectoryPath = "testResultsDirectoryPath"
		expectedValue = `
GATLING_AGGREGATE_DIR=testResultsDirectoryPath
DIR_NAME=$(dirname ${GATLING_AGGREGATE_DIR})
BASE_NAME=$(basename ${GATLING_AGGREGATE_DIR})
gatling.sh -rf ${DIR_NAME} -ro ${BASE_NAME}
`
	})

	It("getExceptValue", func() {
		Expect(GetGatlingGenerateReportCommand(resultsDirectoryPath)).To(Equal(expectedValue))
	})
})

var _ = Describe("GetGatlingTransferReportCommand", func() {
	var (
		resultsDirectoryPath string
		provider             string
		region               string
		storagePath          string
		expectedValue        string
	)

	BeforeEach(func() {
		resultsDirectoryPath = "testResultsDirectoryPath"
		region = "ap-northeast-1"
		storagePath = "testStoragePath"
	})

	Context("Provider is aws", func() {
		BeforeEach(func() {
			provider = "aws"
			expectedValue = `
GATLING_AGGREGATE_DIR=testResultsDirectoryPath
rclone config create s3 s3 env_auth=true region ap-northeast-1
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" --s3-no-check-bucket --s3-env-auth testStoragePath
`
		})
		It("provider is aws", func() {
			Expect(GetGatlingTransferReportCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is gcp", func() {
		BeforeEach(func() {
			provider = "gcp"
			expectedValue = `
GATLING_AGGREGATE_DIR=testResultsDirectoryPath
# assumes gcs bucket using uniform bucket-level access control
rclone config create gs "google cloud storage" bucket_policy_only true --non-interactive
rclone copy ${GATLING_AGGREGATE_DIR} --exclude "*.log" testStoragePath
`
		})
		It("Provider is gcp", func() {
			Expect(GetGatlingTransferReportCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is non-supported one", func() {
		BeforeEach(func() {
			provider = "foo"
			expectedValue = ""
		})
		It("Provide is non-supported one", func() {
			Expect(GetGatlingTransferReportCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})

	Context("Provider is empty", func() {
		BeforeEach(func() {
			provider = ""
			expectedValue = ""
		})
		It("Provider is empty", func() {
			Expect(GetGatlingTransferReportCommand(resultsDirectoryPath, provider, region, storagePath)).To(Equal(expectedValue))
		})
	})
})

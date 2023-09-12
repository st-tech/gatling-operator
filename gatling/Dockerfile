# This is modified based on the original file:
# https://github.com/denvazh/gatling/tree/master/3.2.1
#
# Gatling is a highly capable load testing tool.
#
# Documentation: https://gatling.io/docs/3.2/
# Cheat sheet: https://gatling.io/docs/3.2/cheat-sheet/

FROM public.ecr.aws/docker/library/eclipse-temurin:8-jdk-jammy

# working directory for gatling
WORKDIR /opt

# gating version
ENV GATLING_VERSION 3.2.1
ENV APP_HOME  /opt/gatling
ENV APP_USER  gatling
ENV APP_GROUP appgroup

# install gatling
RUN set -x && mkdir -p gatling && \
  apt-get update && \
  apt-get -y upgrade && \
  apt-get install -y --no-install-recommends curl wget libc6 unzip awscli nginx && \
  mkdir -p /tmp/downloads && \
  wget -q -O /tmp/downloads/gatling-$GATLING_VERSION.zip \
  https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/$GATLING_VERSION/gatling-charts-highcharts-bundle-$GATLING_VERSION-bundle.zip && \
  mkdir -p /tmp/archive && cd /tmp/archive && \
  unzip /tmp/downloads/gatling-$GATLING_VERSION.zip && \
  mv /tmp/archive/gatling-charts-highcharts-bundle-$GATLING_VERSION/* /opt/gatling/ && \
  rm -rf /opt/gatling/user-files/simulations/computerdatabase /tmp/* && \
  addgroup --system ${APP_GROUP} && \
  adduser --system ${APP_USER} --ingroup ${APP_GROUP} --home ${APP_HOME} --no-create-home && \
  chown -R ${APP_USER}:${APP_GROUP} ${APP_HOME} && \
  chown -R ${APP_USER}:${APP_GROUP} /var/log/nginx /var/lib/nginx /usr/share/nginx

USER gatling

# change context to gatling directory
WORKDIR  /opt/gatling

# set directories below to be mountable from host
VOLUME ["/opt/gatling/conf", "/opt/gatling/results", "/opt/gatling/user-files"]

# copy local files to gatling directory
COPY user-files/simulations user-files/simulations
COPY user-files/resources user-files/resources
COPY conf conf

# set environment variables
ENV PATH /opt/gatling/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV GATLING_HOME /opt/gatling

ENTRYPOINT ["gatling.sh"]

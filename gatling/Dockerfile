# This is modified based on the original file:
# https://github.com/denvazh/gatling/tree/master/3.2.1
#
# Gatling is a highly capable load testing tool.

FROM openjdk:21-jdk-slim-bullseye

# create user/group
RUN groupadd -g 1000 gatling && \
  useradd -l -u 1000 -m gatling -g gatling

# working directory for gatling
WORKDIR /opt

# gating version
ENV GATLING_VERSION 3.10.5

# create directory for gatling install
RUN mkdir -p gatling

# install gatling
RUN apt-get update && apt-get upgrade -y && apt-get install -y wget unzip &&  \
  mkdir -p /tmp/downloads && \
  wget -q -O /tmp/downloads/gatling-$GATLING_VERSION.zip \
  https://repo1.maven.org/maven2/io/gatling/highcharts/gatling-charts-highcharts-bundle/$GATLING_VERSION/gatling-charts-highcharts-bundle-$GATLING_VERSION-bundle.zip && \
  mkdir -p /tmp/archive && cd /tmp/archive && \
  unzip /tmp/downloads/gatling-$GATLING_VERSION.zip && \
  mv /tmp/archive/gatling-charts-highcharts-bundle-$GATLING_VERSION/* /opt/gatling/ && \
  rm -rf /opt/gatling/user-files/simulations/computerdatabase /tmp/*  && \
  chown -R gatling:gatling /opt/gatling

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

FROM       quay.io/prometheus/busybox:latest
MAINTAINER Frederick Cai <frederick@movio.co>

RUN mkdir -p /exporter

ADD healthcheck-prometheus-exporter /exporter/
ADD config.yaml /exporter/

WORKDIR /exporter

EXPOSE 8080
ENTRYPOINT  ["/exporter/healthcheck-prometheus-exporter"]

FROM golang:1.13.0-alpine3.10 as builder

ARG COMMIT_SHA
ARG BRANCH_NAME
ARG REPO_NAME

ENV CGO_ENABLED="0" \
    GOOS="linux"

COPY . /work

WORKDIR /work

RUN go test -test.short ./... && \
    go build -v -ldflags "-X main.app=${REPO_NAME} -X main.version=${ESTAFETTE_BUILD_VERSION} -X main.revision=${COMMIT_SHA} -X main.branch=${BRANCH_NAME} -X main.buildDate=${ESTAFETTE_BUILD_DATETIME}"

FROM scratch as runtime

LABEL maintainer="JorritSalverda" \
      description="The tp-link-hs110-bigquery-exporter component is an application that extracts power metering information from TP-Link HS110 power plugs and stores it in BigQuery"

COPY --from=builder /work/tp-link-hs110-bigquery-exporter /tp-link-hs110-bigquery-exporter
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["tp-link-hs110-bigquery-exporter"]
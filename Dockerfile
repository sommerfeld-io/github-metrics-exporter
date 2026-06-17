FROM golang:1.26.4-alpine3.24 AS build
LABEL maintainer="sebastian@sommerfeld.io"

RUN apk add --no-cache curl git \
    && sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY .git /workspaces/github-metrics-exporter/.git
WORKDIR /workspaces/github-metrics-exporter/src
COPY src /workspaces/github-metrics-exporter/src

## see src/taskfile.yml
RUN task build




FROM build AS acceptance-test
WORKDIR /workspaces/github-metrics-exporter
COPY tests /workspaces/github-metrics-exporter/tests
WORKDIR /workspaces/github-metrics-exporter/tests/acceptance
RUN go mod download && go test -v ./...




FROM alpine:3.24.0 AS run
LABEL maintainer="sebastian@sommerfeld.io"

ARG USER_NAME=ghme
ARG USER_ID=1000
ARG GROUP_ID=1000

RUN addgroup -g ${GROUP_ID} ${USER_NAME} \
    && adduser -D -u ${USER_ID} -G ${USER_NAME} -h /opt/github-metrics-exporter ${USER_NAME} \
    && chown -R ${USER_NAME}:${USER_NAME} /opt/github-metrics-exporter

WORKDIR /opt/github-metrics-exporter
COPY --from=build /workspaces/github-metrics-exporter/src/github-metrics-exporter ./github-metrics-exporter

USER ${USER_NAME}

CMD ["./github-metrics-exporter"]

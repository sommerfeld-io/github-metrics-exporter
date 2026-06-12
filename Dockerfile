FROM golang:1.25.11-alpine3.24 AS build
LABEL maintainer="sebastian@sommerfeld.io"

RUN apk add --no-cache curl \
    && sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

WORKDIR /src
COPY src /src

RUN task build


FROM alpine:3.24.0 AS run
LABEL maintainer="sebastian@sommerfeld.io"

ARG USER_NAME=ghme
ARG USER_ID=1000
ARG GROUP_ID=1000

WORKDIR /opt/github-metrics-exporter

RUN addgroup -g ${GROUP_ID} ${USER_NAME} \
    && adduser -D -u ${USER_ID} -G ${USER_NAME} -h /opt/github-metrics-exporter ${USER_NAME} \
    && chown -R ${USER_NAME}:${USER_NAME} /opt/github-metrics-exporter

COPY --from=build /src/github-metrics-exporter ./github-metrics-exporter

USER ${USER_NAME}

CMD ["./github-metrics-exporter"]

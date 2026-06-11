FROM golang:1.25.11-alpine3.24 AS build
LABEL maintainer="sebastian@sommerfeld.io"

RUN apk add --no-cache curl \
    && sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

WORKDIR /src
COPY src /src

RUN task build


FROM alpine:3.24.0 AS run
LABEL maintainer="sebastian@sommerfeld.io"

WORKDIR /opt/github-metrics-exporter
COPY --from=build /src/github-metrics-exporter ./github-metrics-exporter
CMD ["./github-metrics-exporter"]

# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG BININFO_VERSION
ARG BININFO_COMMIT_HASH
ARG BININFO_BUILD_DATE

RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
        -X github.com/sapcc/go-api-declarations/bininfo.binName=oomkill-exporter \
        -X github.com/sapcc/go-api-declarations/bininfo.version=${BININFO_VERSION} \
        -X github.com/sapcc/go-api-declarations/bininfo.commit=${BININFO_COMMIT_HASH} \
        -X github.com/sapcc/go-api-declarations/bininfo.buildDate=${BININFO_BUILD_DATE}" \
    -o oomkill-exporter \
    ./cmd/oomkill-exporter

FROM alpine:3.23

RUN apk upgrade --no-cache && \
    apk del --no-cache apk-tools alpine-keys

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/oomkill-exporter /usr/bin/

RUN /usr/bin/oomkill-exporter --version

ENTRYPOINT ["/usr/bin/oomkill-exporter"]

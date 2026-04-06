# SPDX-FileCopyrightText: (C) 2026 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.26.1@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 AS base

ENV CGO_ENABLED=0

# Set proxy environment variables using build arguments
ARG http_proxy
ARG https_proxy
ARG no_proxy

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# rebuild atlas from source to match go 1.26.1 and patch CVE-affected deps
# hash for atlas main repo
ENV ATLAS_HASH=6243aa41c3eea0dbe7c611a2eeb249ed7ef07670
WORKDIR /build/src
RUN git clone https://github.com/ariga/atlas.git
WORKDIR /build/src/atlas
RUN git checkout $ATLAS_HASH
WORKDIR /build/src/atlas/cmd/atlas
# Upgrade vulnerable indirect deps in cmd/atlas/go.mod before building:
#   filippo.io/edwards25519 v1.1.0 -> v1.2.0 (fixes CVE-2026-33186)
#   go.opentelemetry.io/otel/sdk v1.37.0 -> v1.43.0 (fixes CVE-2026-24051)
#   google.golang.org/grpc -> v1.79.3 (fixes CVE-2026-33186 grpc auth bypass)
RUN sed -i 's/go 1\.[0-9]\+\.[0-9]\+/go 1.26.1/g' go.mod && \
    go get filippo.io/edwards25519@v1.2.0 && \
    go get go.opentelemetry.io/otel/sdk@v1.43.0 && \
    go get google.golang.org/grpc@v1.79.3 && \
    go mod tidy && \
    go build -o /go/bin/atlas .

WORKDIR /build
RUN mkdir -p "$APP_ROOT/" && mkdir -p /build/etc

ENV APP_ROOT=/build/src/github.com/open-edge-platform/app-orch-catalog

COPY ./api/spec/openapi.yaml /build/etc

COPY ./go.mod ./go.sum $APP_ROOT/
COPY ./vendor $APP_ROOT/vendor

FROM base AS rest-proxy
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p "$APP_ROOT/cmd/rest-proxy" && mkdir -p "$APP_ROOT/internal/version" && mkdir -p "$APP_ROOT/internal/file-handler"

COPY ./cmd/rest-proxy $APP_ROOT/cmd/rest-proxy
COPY ./internal/restproxy $APP_ROOT/internal/restproxy
COPY ./internal/shared $APP_ROOT/internal/shared
COPY ./pkg $APP_ROOT/pkg

WORKDIR $APP_ROOT

ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_ref=unknown
ARG org_opencord_vcs_dirty=unknown

RUN GO111MODULE=on go version && go build -mod=vendor -o /build/bin/rest-proxy \
        -gcflags="all=-spectre=all -N -l" -asmflags="-spectre=all" --trimpath \
        -ldflags "all=-s -w" -ldflags \
        "-X github.com/open-edge-platform/app-orch-catalog/internal/version.Version=$org_label_schema_version \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitCommit=$org_label_schema_vcs_ref  \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitDirty=$org_opencord_vcs_dirty \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GoVersion=$(go version 2>&1 | sed -E  's/^go version go//' | sed -E 's/ .*$//') \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.Os=$(go env GOHOSTOS) \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.Arch=$(go env GOHOSTARCH)" \
         ./cmd/rest-proxy

FROM base AS application-catalog
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir -p "$APP_ROOT/cmd/application-catalog" && mkdir -p "$APP_ROOT/cmd/healthcheck"
COPY ./cmd/application-catalog $APP_ROOT/cmd/application-catalog
COPY ./cmd/healthcheck $APP_ROOT/cmd/healthcheck
COPY ./pkg $APP_ROOT/pkg
COPY ./internal $APP_ROOT/internal

WORKDIR $APP_ROOT

ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_ref=unknown
ARG org_opencord_vcs_dirty=unknown

RUN GO111MODULE=on go version && go build -mod=vendor -o /build/bin/application-catalog \
        -gcflags="all=-spectre=all -N -l" -asmflags="-spectre=all" --trimpath \
        -ldflags "all=-s -w" -ldflags \
        "-X github.com/open-edge-platform/app-orch-catalog/internal/version.Version=$org_label_schema_version \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitCommit=$org_label_schema_vcs_ref  \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitDirty=$org_opencord_vcs_dirty \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GoVersion=$(go version 2>&1 | sed -E  's/^go version go//' | sed -E 's/ .*$//') \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.Os=$(go env GOHOSTOS) \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.Arch=$(go env GOHOSTARCH)" \
         ./cmd/application-catalog

RUN go build -mod=vendor -o /build/bin/healthcheck --trimpath \
        -ldflags "-s -w" \
        ./cmd/healthcheck

FROM gcr.io/distroless/static:nonroot@sha256:e3f945647ffb95b5839c07038d64f9811adf17308b9121d8a2b87b6a22a80a39
USER nonroot

ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_url=unknown
ARG org_label_schema_vcs_ref=unknown

LABEL org.opencontainers.image.version=$org_label_schema_version \
      org.opencontainers.image.source=$org_label_schema_vcs_url \
      org.opencontainers.image.revision=$org_label_schema_vcs_ref

USER nobody

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

COPY --from=rest-proxy /build/bin/rest-proxy /usr/local/bin/

COPY --from=application-catalog /build/bin/application-catalog /usr/local/bin/
COPY --from=application-catalog /build/bin/healthcheck /usr/local/bin/
COPY --from=application-catalog /build/etc/openapi.yaml /usr/local/etc/openapi.yaml

# Copy database migrations
COPY --from=application-catalog --chown=nobody:nobody /build/src/github.com/open-edge-platform/app-orch-catalog/internal/ent/migrate/migrations /usr/share/migrations

# Copy atlas tool
COPY --from=base --chown=nobody:nobody /go/bin/atlas /usr/local/bin/atlas

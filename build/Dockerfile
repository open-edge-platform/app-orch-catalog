# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24.2@sha256:d9db32125db0c3a680cfb7a1afcaefb89c898a075ec148fdc2f0f646cc2ed509 AS base

ENV CGO_ENABLED=0

# Set proxy environment variables using build arguments
ARG http_proxy
ARG https_proxy
ARG no_proxy

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ENV ATLAS_VERSION=v0.32.1-9f69a85-canary
ENV ATLAS_GO=/tmp/atlasgo.sh
ENV ATLAS_URL=https://atlasgo.sh
RUN curl -sSf $ATLAS_URL > $ATLAS_GO
RUN chmod +x $ATLAS_GO && $ATLAS_GO -y -o /go/bin/atlas
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
        -gcflags="all=-spectre=all -N -l" -asmflags="all=-spectre=all" --trimpath \
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
RUN mkdir -p "$APP_ROOT/cmd/application-catalog"
COPY ./cmd/application-catalog $APP_ROOT/cmd/application-catalog
COPY ./pkg $APP_ROOT/pkg
COPY ./internal $APP_ROOT/internal

WORKDIR $APP_ROOT

ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_ref=unknown
ARG org_opencord_vcs_dirty=unknown

RUN GO111MODULE=on go version && go build -mod=vendor -o /build/bin/application-catalog \
        -gcflags="all=-spectre=all -N -l" -asmflags="all=-spectre=all" --trimpath \
        -ldflags "all=-s -w" -ldflags \
        "-X github.com/open-edge-platform/app-orch-catalog/internal/version.Version=$org_label_schema_version \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitCommit=$org_label_schema_vcs_ref  \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitDirty=$org_opencord_vcs_dirty \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.GoVersion=$(go version 2>&1 | sed -E  's/^go version go//' | sed -E 's/ .*$//') \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.Os=$(go env GOHOSTOS) \
         -X github.com/open-edge-platform/app-orch-catalog/internal/version.Arch=$(go env GOHOSTARCH)" \
         ./cmd/application-catalog

FROM gcr.io/distroless/static:nonroot@sha256:c0f429e16b13e583da7e5a6ec20dd656d325d88e6819cafe0adb0828976529dc
USER nonroot

ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_url=unknown
ARG org_label_schema_vcs_ref=unknown

LABEL org.opencontainers.image.version=$org_label_schema_version \
      org.opencontainers.image.source=$org_label_schema_vcs_url \
      org.opencontainers.image.revision=$org_label_schema_vcs_ref

USER nobody

COPY --from=rest-proxy /build/bin/rest-proxy /usr/local/bin/

COPY --from=application-catalog /build/bin/application-catalog /usr/local/bin/
COPY --from=application-catalog /build/etc/openapi.yaml /usr/local/etc/openapi.yaml

# Copy database migrations
COPY --from=application-catalog --chown=nobody:nobody /build/src/github.com/open-edge-platform/app-orch-catalog/internal/ent/migrate/migrations /usr/share/migrations

# Copy atlas tool
COPY --from=base --chown=nobody:nobody /go/bin/atlas /usr/local/bin/atlas

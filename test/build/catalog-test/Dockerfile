# SPDX-FileCopyrightText: 2023-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.22.5 AS build
ENV APP_ROOT=/build/src/github.com/open-edge-platform/app-orch-catalog

COPY ./go.mod ./go.sum $APP_ROOT/
COPY ./vendor $APP_ROOT/vendor
COPY ./test/vendor $APP_ROOT/test/vendor

COPY ./test $APP_ROOT/test
COPY ./pkg $APP_ROOT/pkg

WORKDIR $APP_ROOT
RUN cd $APP_ROOT/test && GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -mod=vendor -o $APP_ROOT/build/_output/catalog-test -v -ldflags \
            "-X github.com/open-edge-platform/app-orch-catalog/internal/version.Version=$org_label_schema_version \
             -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitCommit=$org_label_schema_vcs_ref  \
             -X github.com/open-edge-platform/app-orch-catalog/internal/version.GitDirty=$org_opencord_vcs_dirty \
             -X github.com/open-edge-platform/app-orch-catalog/internal/version.GoVersion=$(go version 2>&1 | sed -E  's/^go version go//' | sed -E 's/ .*$//') \
             -X github.com/open-edge-platform/app-orch-catalog/internal/version.Os=$(go env GOHOSTOS) \
             -X github.com/open-edge-platform/app-orch-catalog/internal/version.Arch=$(go env GOHOSTARCH)" \
      ./cmd/catalog-test

FROM scratch
ENV APP_ROOT=/build/src/github.com/open-edge-platform/app-orch-catalog

COPY --from=build $APP_ROOT/build/_output/catalog-test /catalog-test
COPY --from=build $APP_ROOT/test/basic/1x1.png /test/basic/1x1.png
COPY --from=build $APP_ROOT/test/basic/testdata /test/basic/testdata
ENTRYPOINT ["/catalog-test"]

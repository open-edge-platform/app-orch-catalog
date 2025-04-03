<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->
Developer Guide
===============

# Tooling
On MacOS or Linux you can use the [asdf](https://asdf-vm.com/guide/getting-started.html) tool to install many
of the tools needed for development.

> ASDF requires plugins for each tool listed - see https://github.com/asdf-vm/asdf-plugins - `asdf plugin add <name> <url>`

The [.tool-versions](../.tool-versions) file contains a list of tools and versions that can be installed with
```shell
asdf install
```

Additionally, to generate code from Protobuf some more tools are needed:

```shell
go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
```

# Object Model
The object model shown in the [architecture](architecture.md) is implemented by 2 models:

* an [ENT schema](../internal/ent/schema)
* a [protobuf definition](../api/catalog/v2/resources.proto)

The application-catalog itself is started from `cmd/grpc/main.go`, and runs a gRPC server on Port 8080 
and connects to a Postgres database on the backend on Port 5432.

A 2nd executable provides a `REST` front end and is started with `cmd/proxy/main.go`, and
is exposed on Port 8081. 

## ENT Schema
The schema describes all the objects and their relations.

After changing anything in `internal/ent/schema` regenerate the code:
```shell
make ent-generate
```

To see the schema summary
```shell
make ent-describe
```
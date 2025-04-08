<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Developer Guide

## Tooling

On macOS or Linux, you can use the [asdf](https://asdf-vm.com/guide/getting-started.html) tool to install many
of the tools needed for development.

> ASDF requires plugins for each tool listed - see <https://github.com/asdf-vm/asdf-plugins> - `asdf plugin add <name> <url>`

The [.tool-versions](../.tool-versions) file contains a list of tools and versions that can be installed with:

```shell
asdf install
```

Additionally, to generate code from Protobuf, some more tools are needed.
To install them, run the following commands:

```shell
make install-protoc-plugins
make verify-protoc-plugins
```

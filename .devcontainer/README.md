<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Dev Container for App Catalog

There are many ways to use this **Devcontainer**

- VS Code Remote Containers. https://code.visualstudio.com/docs/devcontainers/create-dev-container
- IntelliJ IDEA Remote Development. https://www.jetbrains.com/help/idea/connect-to-devcontainer.html
- devcontainer CLI. https://github.com/devcontainers/cli?tab=readme-ov-file#npm-install

## Command line usage

To run the devcontainer CLI, you can use the following command:

```bash
devcontainer up --workspace-folder .
```

To run any of the make targets inside the Devcontainer, you can use the following command for example:

```bash
devcontainer exec --workspace-folder . make lint
```

If you need to rebuild the devcontainer, you can use the following command:

```bash
devcontainer up --remove-existing-container --workspace-folder .
```
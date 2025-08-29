<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Upgraded Deployment Package for Cirros VM

This Deployment Package is intended to use for testing upgrades.

It is identical to the one in ../cirros-vm with the sole difference that is uses version 1.0.0 of the cirros vm image rather than 0.36.4

For documentation, see the README in ../cirros-vm

# Testing upgrades

To test an upgrade, follow these steps:

1. Import version 0.1.0 of the Cirros VM to the orchestrator

2. Deploy version 0.1.0 of the Cirros VM

3. Use VNC to inspect that the Cirros VM is alive and health. `uname -a` should show kernel 4.4.0

4. Import version 0.1.1 of the Cirros VM to the orchestrator

5. Orchestrator should show "upgrades available" on the VM deployed in step 2.

6. Upgrade the VM from 0.1.0 to 0.1.1 using the orchestrator. The VM will be unavailable for a short while while it restarts. This may take a few minutes.

7. Use VNC to inspect that the Cirros VM is alive and health. `uname -a` should show kernel 5.3.0, confirming the VM has been upgraded.

<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Deployment Package for Cirros VM

This is a deployment package for Cirros VM. It is a lightweight Linux distribution designed for testing and development purposes. 
The package includes a Helm chart that deploys the Cirros VM image and configures it to run on the Open Edge Platform.

## Importing Deployment Package
All the files including [Application](app.yaml), [Deployment Package](dp.yaml), [Registry](registry.yaml), and [Overriding values](values.yaml) 
should be imported to Edge Manageability Framework (EMF) using the EMF Web UI. 
For more information, refer to [Import Deployment Packages](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/package_software/import_deployment.html) 

## Deployment

#### Prerequisites
The Virtualization Extension package, which is preloaded, 
needs to be installed using the `Software Emulation Configuration` profile before installing the Cirros VM deployment package.

The Cirros VM deployment package is designed to be used on Edge Manageability Framework (EMF). 
After importing the package, you can deploy it using the EMF Web UI. 
For more information, refer to [Setup A Deployment](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/package_software/setup_deploy.html)

After deploying virtualization extension and Cirros VM, you can perform life cycle operations on the Cirros VM using the EMF Web UI.
For more information, refer to [Perform Actions on the Virtual Machines](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/package_software/vm_actions.html)


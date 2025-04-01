# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

import future.keywords.in

CreateDeploymentPackageRequest {
    hasWriteAccess
}

UpdateDeploymentPackageRequest {
    hasWriteAccess
}

DeleteDeploymentPackageRequest {
    hasWriteAccess
}

GetDeploymentPackageRequest {
    hasReadAccess
}

GetDeploymentPackageVersionsRequest {
    hasReadAccess
}

ListDeploymentPackagesRequest {
    hasReadAccess
}

WatchDeploymentPackagesRequest {
    hasReadAccess
}
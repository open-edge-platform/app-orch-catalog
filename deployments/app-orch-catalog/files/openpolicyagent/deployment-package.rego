# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

import future.keywords.in

CreateDeploymentPackageRequest if {
    hasWriteAccess
}

UpdateDeploymentPackageRequest if {
    hasWriteAccess
}

DeleteDeploymentPackageRequest if {
    hasWriteAccess
}

GetDeploymentPackageRequest if {
    hasReadAccess
}

GetDeploymentPackageVersionsRequest if {
    hasReadAccess
}

ListDeploymentPackagesRequest if {
    hasReadAccess
}

WatchDeploymentPackagesRequest if {
    hasReadAccess
}
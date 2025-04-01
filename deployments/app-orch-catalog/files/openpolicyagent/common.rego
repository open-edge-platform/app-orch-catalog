# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

import future.keywords.in

# Allows management of all objects
hasWriteAccess {
    projectRole := sprintf("%s_cat-rw", [input.metadata.activeprojectid[0]])
    some role in input.metadata["realm_access/roles"] # iteration
    [projectRole][_] == role
}
# OR for m2m there is a non-project specific role
hasWriteAccess {
    some role in input.metadata["realm_access/roles"] # iteration
    ["ao-m2m-rw"][_] == role
}

# This is used for access to all objects
hasReadAccess {
    projectRole := sprintf("%s_cat-r", [input.metadata.activeprojectid[0]])
    some role in input.metadata["realm_access/roles"] # iteration
    [projectRole][_] == role
}
# OR with new short role names rw includes read access
hasReadAccess {
    projectRole := sprintf("%s_cat-rw", [input.metadata.activeprojectid[0]])
    some role in input.metadata["realm_access/roles"] # iteration
    [projectRole][_] == role
}
# OR for m2m there is a non-project specific role
hasReadAccess {
    some role in input.metadata["realm_access/roles"] # iteration
    ["ao-m2m-rw"][_] == role
}

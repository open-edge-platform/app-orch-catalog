# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

CreateApplicationRequest {
    hasWriteAccess
}

UpdateApplicationRequest {
    hasWriteAccess
}

DeleteApplicationRequest {
    hasWriteAccess
}

GetApplicationRequest {
    hasReadAccess
}

GetApplicationVersionsRequest {
    hasReadAccess
}

GetApplicationReferenceCountRequest {
    hasReadAccess
}

ListApplicationsRequest {
    hasReadAccess
}

WatchApplicationsRequest {
    hasReadAccess
}
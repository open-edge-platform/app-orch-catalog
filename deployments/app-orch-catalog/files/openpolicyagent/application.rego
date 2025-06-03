# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

CreateApplicationRequest if {
    hasWriteAccess
}

UpdateApplicationRequest if {
    hasWriteAccess
}

DeleteApplicationRequest if {
    hasWriteAccess
}

GetApplicationRequest if {
    hasReadAccess
}

GetApplicationVersionsRequest if {
    hasReadAccess
}

GetApplicationReferenceCountRequest if {
    hasReadAccess
}

ListApplicationsRequest if {
    hasReadAccess
}

WatchApplicationsRequest if {
    hasReadAccess
}
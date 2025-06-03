# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

CreateArtifactRequest if {
    hasWriteAccess
}

UpdateArtifactRequest if {
    hasWriteAccess
}

DeleteArtifactRequest if {
    hasWriteAccess
}

GetArtifactRequest if {
    hasReadAccess
}

ListArtifactsRequest if {
    hasReadAccess
}

WatchArtifactsRequest if {
    hasReadAccess
}
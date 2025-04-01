# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

CreateArtifactRequest {
    hasWriteAccess
}

UpdateArtifactRequest {
    hasWriteAccess
}

DeleteArtifactRequest {
    hasWriteAccess
}

GetArtifactRequest {
    hasReadAccess
}

ListArtifactsRequest {
    hasReadAccess
}

WatchArtifactsRequest {
    hasReadAccess
}
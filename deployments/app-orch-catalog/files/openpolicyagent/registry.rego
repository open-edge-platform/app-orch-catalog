# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

CreateRegistryRequest {
    hasWriteAccess
}

UpdateRegistryRequest {
    hasWriteAccess
}

DeleteRegistryRequest {
    hasWriteAccess
}

GetRegistryWithSensitiveInfoRequest {
    hasReadAccess
}

GetRegistryRequest {
    hasReadAccess
}

ListRegistriesWithSensitiveInfoRequest {
    hasReadAccess
}

ListRegistriesRequest {
    hasReadAccess
}

WatchRegistriesWithSensitiveInfoRequest {
    hasReadAccess
}

WatchRegistriesRequest {
    hasReadAccess
}

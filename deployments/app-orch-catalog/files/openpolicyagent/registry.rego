# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

CreateRegistryRequest if {
    hasWriteAccess
}

UpdateRegistryRequest if {
    hasWriteAccess
}

DeleteRegistryRequest if {
    hasWriteAccess
}

GetRegistryWithSensitiveInfoRequest if {
    hasReadAccess
}

GetRegistryRequest if {
    hasReadAccess
}

ListRegistriesWithSensitiveInfoRequest if {
    hasReadAccess
}

ListRegistriesRequest if {
    hasReadAccess
}

WatchRegistriesWithSensitiveInfoRequest if {
    hasReadAccess
}

WatchRegistriesRequest if {
    hasReadAccess
}

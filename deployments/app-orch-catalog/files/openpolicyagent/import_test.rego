# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

# test ImportRequest rule as write role - ALLOWED
test_import_write_role if {
    ImportRequest with input as {
        "request": {
        },
        "metadata": {
            "activeprojectid": [
            "2724b4fc-745e-4537-b76c-13907a9ea831"
            ],
            "client": [
            "catalog-cli"
            ],
            "realm_access/roles": [
            "default-roles-master",
            "offline_access",
            "2724b4fc-745e-4537-b76c-13907a9ea831_cat-rw",
            "uma_authorization"
            ]
        }
    }
}

# test ImportRequest rule as read only role - DENIED
test_import_read_role if {
    not ImportRequest with input as {
        "request": {
        },
        "metadata": {
            "activeprojectid": [
            "2724b4fc-745e-4537-b76c-13907a9ea831"
            ],
            "client": [
            "catalog-cli"
            ],
            "realm_access/roles": [
            "default-roles-master",
            "offline_access",
            "2724b4fc-745e-4537-b76c-13907a9ea831_cat-r",
            "uma_authorization"
            ]
        }
    }
}
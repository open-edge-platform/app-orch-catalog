# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

# test CreateRegistryRequest rule as write role - ALLOWED
test_create_registry_write_role if {
    CreateRegistryRequest with input as {
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

# test DeleteRegistryRequest rule as read-only - DENIED
test_delete_registry_read_role if {
    not DeleteRegistryRequest with input as {
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

# test DeleteRegistryRequest rule as write role - ALLOWED
test_delete_registry_write_role if {
    DeleteRegistryRequest with input as {
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

# test GetRegistryRequest rule as restricted read role - DENIED
test_get_registry_restricted_read_role if {
    not GetRegistryRequest with input as {
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
          "2724b4fc-745e-4537-b76c-13907a9ea831_tc-r",
          "uma_authorization"
        ]
      }
   }
}

# test GetArtifactRequest rule as read role - ALLOWED
test_get_registry_read_role if {
    GetRegistryRequest with input as {
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

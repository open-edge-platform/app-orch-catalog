# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

# test CreateArtifactRequest rule as write role - ALLOWED
test_create_artifact_write_role if {
    CreateArtifactRequest with input as {
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

# test CreateArtifactRequest rule as read-only - DENIED
test_create_artifact_read_role if {
    not CreateArtifactRequest with input as {
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
          "2724b4fc-745e-4537-b76c-13907a9ea831_default-roles-master",
          "2724b4fc-745e-4537-b76c-13907a9ea831_cat-r",
          "offline_access",
          "uma_authorization"
        ]
      }
    }
}

# test GetArtifactRequest rule as read role - ALLOWED
test_get_artifact_read_role if {
    GetArtifactRequest with input as {
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
          "2724b4fc-745e-4537-b76c-13907a9ea831_default-roles-master",
          "2724b4fc-745e-4537-b76c-13907a9ea831_cat-r",
          "offline_access",
          "uma_authorization"
        ]
      }
  }
}

# test GetArtifactRequest rule as restricted read role - DENIED
test_get_artifact_restricted_read_role if {
    not GetArtifactRequest with input as {
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
          "uma_authorization"
        ]
      }
  }
}
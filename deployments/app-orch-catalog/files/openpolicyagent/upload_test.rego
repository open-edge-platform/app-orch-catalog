# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

# test UploadEntitiesRequest rule as all required write roles - ALLOWED
test_upload_entities_write_role if {
    UploadCatalogEntitiesRequest with input as {
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

# test UploadEntitiesRequest rule as read role - DENIED
test_upload_entities_read_role if {
    not UploadCatalogEntitiesRequest with input as {
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

# test UploadEntitiesRequest rule as read role - DENIED
test_upload_entities_read_role if {
    not UploadCatalogEntitiesRequest with input as {
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

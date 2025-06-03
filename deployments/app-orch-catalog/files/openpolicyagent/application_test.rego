# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package catalogv3

# test CreateApplicationRequest rule as write role - ALLOWED
test_create_application_write_role if {
    CreateApplicationRequest with input as {
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

# test CreateApplicationRequest rule as read role - DENIED
test_create_application_read_role if {
    not CreateApplicationRequest with input as {
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

# test UpdateApplicationRequest rule as write role - ALLOWED
test_update_application_write_role if {
    UpdateApplicationRequest with input as {
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

# test UpdateApplicationRequest rule as read role - DENIED
test_update_application_read_role if {
    not UpdateApplicationRequest with input as {
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

# test DeleteApplicationRequest rule as write role - ALLOWED
test_delete_application_write_role if {
    DeleteApplicationRequest with input as {
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

# test DeleteApplicationRequest rule as read role - DENIED
test_delete_application_read_role if {
    not DeleteApplicationRequest with input as {
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

# test GetApplicationRequest rule as read role - ALLOWED
test_get_application_read_role if {
    GetApplicationRequest with input as {
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

# test GetApplicationVersionsRequest rule as read role - ALLOWED
test_get_application_versions_read_role if {
    GetApplicationVersionsRequest with input as {
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

# test GetApplicationReferenceCountRequest rule as read role - ALLOWED
test_get_application_reference_count_read_role if {
    GetApplicationReferenceCountRequest with input as {
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

# test ListApplicationsRequest rule as read role - ALLOWED
test_list_application_read_role if {
    ListApplicationsRequest with input as {
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

# test WatchApplicationsRequest rule as read role - ALLOWED
test_watch_application_read_role if {
    WatchApplicationsRequest with input as {
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
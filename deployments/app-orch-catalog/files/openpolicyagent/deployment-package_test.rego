


package catalogv3

# test CreateDeploymentPackageRequest rule as write role - ALLOWED
test_create_deployment_package_write_role if {
    CreateDeploymentPackageRequest with input as {
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

# test CreateDeploymentPackageRequest rule as write role for m2m account - ALLOWED
test_create_deployment_package_read_role_m2m if {
    CreateDeploymentPackageRequest with input as {
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
          "ao-m2m-rw",
          "uma_authorization"
        ]
      }
  }
}

# test CreateDeploymentPackageRequest rule as old write role - DENIED
test_create_deployment_package_read_role_old_denied if {
    not CreateDeploymentPackageRequest with input as {
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
          "2724b4fc-745e-4537-b76c-13907a9ea831_catalog-other-write-role",
          "uma_authorization"
        ]
      }
  }
}

# test CreateDeploymentPackageRequest rule as write role m2m - DENIED
test_create_deployment_package_read_role if {
    not CreateDeploymentPackageRequest with input as {
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
          "catalog-restricted-write-role",
          "uma_authorization"
        ]
      }
  }
}

# test CreateDeploymentPackageRequest rule as read role - DENIED
test_create_deployment_package_read_role if {
    not CreateDeploymentPackageRequest with input as {
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

# test CreateDeploymentPackageRequest rule as read role old - DENIED
test_create_deployment_package_read_role_old if {
    not CreateDeploymentPackageRequest with input as {
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
          "2724b4fc-745e-4537-b76c-13907a9ea831_catalog-other-read-role",
          "uma_authorization"
        ]
      }
  }
}

# test GetDeploymentPackageRequest rule as read role - ALLOWED
test_get_deployment_package_read_role if {
    GetDeploymentPackageRequest with input as {
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

# test GetDeploymentPackageRequest rule as read-write role - ALLOWED
test_get_deployment_package_read_write_role if {
    GetDeploymentPackageRequest with input as {
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

# test GetDeploymentPackageRequest rule as read role old - DENIED
test_get_deployment_package_read_role_m2m if {
    not GetDeploymentPackageRequest with input as {
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
          "2724b4fc-745e-4537-b76c-13907a9ea831_catalog-other-read-role",
          "uma_authorization"
        ]
      }
  }
}

# test GetDeploymentPackageRequest rule as read role m2m - ALLOWED
test_get_deployment_package_read_role_m2m if {
    GetDeploymentPackageRequest with input as {
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
          "ao-m2m-rw",
          "uma_authorization"
        ]
      }
  }
}

# test GetDeploymentPackageRequest rule as read role - DENIED
test_get_deployment_package_read_role_old_m2m if {
    not GetDeploymentPackageRequest with input as {
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
          "catalog-other-read-role",
          "uma_authorization"
        ]
      }
  }
}

# test GetDeploymentPackageRequest rule as no role - DENIED
test_get_deployment_package_read_role_none if {
    not GetDeploymentPackageRequest with input as {
      "request": {
      },
      "metadata": {
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

# test GetDeploymentPackageVersionsRequest versions rule as read role - ALLOWED
test_get_deployment_package_versions_read_role if {
    GetDeploymentPackageVersionsRequest with input as {
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

# test ListDeploymentPackageRequest rule as read-only role - ALLOWED
test_list_deployment_package_read_role if {
    ListDeploymentPackagesRequest with input as {
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

# test UpdateDeploymentPackageRequest rule as read-write role - ALLOWED
test_update_deployment_package_read_role if {
    UpdateDeploymentPackageRequest with input as {
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

# test UpdateDeploymentPackageRequest rule as read-only role - DENIED
test_update_deployment_package_read_role if {
    not UpdateDeploymentPackageRequest with input as {
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

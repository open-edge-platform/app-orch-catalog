-- Modify "applications" table
ALTER TABLE "applications" DROP CONSTRAINT "applications_publishers_applications", DROP COLUMN "publisher_applications", ADD COLUMN "project_uuid" character varying NOT NULL DEFAULT 'default';
-- Create index "application_project_uuid_name_version" to table: "applications"
CREATE UNIQUE INDEX "application_project_uuid_name_version" ON "applications" ("project_uuid", "name", "version");
-- Modify "artifacts" table
ALTER TABLE "artifacts" DROP CONSTRAINT "artifacts_publishers_artifacts", DROP COLUMN "publisher_artifacts", ADD COLUMN "project_uuid" character varying NOT NULL DEFAULT 'default';
-- Create index "artifact_project_uuid_name" to table: "artifacts"
CREATE UNIQUE INDEX "artifact_project_uuid_name" ON "artifacts" ("project_uuid", "name");
-- Modify "deployment_packages" table
ALTER TABLE "deployment_packages" DROP CONSTRAINT "deployment_packages_publishers_deployment_packages", DROP COLUMN "publisher_deployment_packages", ADD COLUMN "project_uuid" character varying NOT NULL DEFAULT 'default';
-- Create index "deploymentpackage_project_uuid_name_version" to table: "deployment_packages"
CREATE UNIQUE INDEX "deploymentpackage_project_uuid_name_version" ON "deployment_packages" ("project_uuid", "name", "version");
-- Modify "parameter_templates" table
ALTER TABLE "parameter_templates" ADD COLUMN "mandatory" boolean NULL, ADD COLUMN "secret" boolean NULL;
-- Modify "registries" table
ALTER TABLE "registries" DROP CONSTRAINT "registries_publishers_registries", DROP COLUMN "publisher_registries", ADD COLUMN "project_uuid" character varying NOT NULL DEFAULT 'default';
-- Create index "registry_project_uuid_name" to table: "registries"
CREATE UNIQUE INDEX "registry_project_uuid_name" ON "registries" ("project_uuid", "name");

DROP TABLE "publishers";
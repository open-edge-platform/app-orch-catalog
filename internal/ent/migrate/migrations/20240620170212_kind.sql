-- Modify "applications" table
ALTER TABLE "applications" ADD COLUMN "kind" character varying NULL;
-- Modify "deployment_packages" table
ALTER TABLE "deployment_packages" ADD COLUMN "kind" character varying NULL;

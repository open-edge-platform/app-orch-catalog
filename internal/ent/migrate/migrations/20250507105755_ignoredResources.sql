-- Modify "ignored_resources" table
ALTER TABLE "ignored_resources" ALTER COLUMN "namespace" SET NOT NULL;
-- Create index "ignoredresource_name_kind_name_c895d252baa82b4b02650b3d765e507a" to table: "ignored_resources"
CREATE UNIQUE INDEX "ignoredresource_name_kind_name_c895d252baa82b4b02650b3d765e507a" ON "ignored_resources" ("name", "kind", "namespace", "application_ignored_resources");
DROP INDEX ignoredresource_name_kind_application_ignored_resources;

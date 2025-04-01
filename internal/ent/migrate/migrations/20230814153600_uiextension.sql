-- Modify "extensions" table
ALTER TABLE "extensions" ADD COLUMN "ui_description" character varying NULL, ADD COLUMN "ui_file_name" character varying NULL, ADD COLUMN "ui_app_name" character varying NULL, ADD COLUMN "ui_module_name" character varying NULL;

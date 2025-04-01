-- Modify "deployment_requirements" table
ALTER TABLE "deployment_requirements" DROP CONSTRAINT "deployment_requirements_profiles_deployment_requirements", ADD CONSTRAINT "deployment_requirements_profiles_deployment_requirements" FOREIGN KEY ("profile_deployment_requirements") REFERENCES "profiles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

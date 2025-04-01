-- Modify "deployment_profiles" table
ALTER TABLE "deployment_profiles" DROP CONSTRAINT "deployment_profiles_deployment_fce10a7872a343d005bd76050d0329a0";
-- Modify "deployment_requirements" table
ALTER TABLE "deployment_requirements" ADD COLUMN "deployment_requirement_deployment_profile_fk" bigint NULL, ADD CONSTRAINT "deployment_requirements_deploy_b543d59ddb0c891b15f0d56140dcf00a" FOREIGN KEY ("deployment_requirement_deployment_profile_fk") REFERENCES "deployment_profiles" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

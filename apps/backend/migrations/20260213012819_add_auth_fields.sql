-- Modify "user_profiles" table
ALTER TABLE "public"."user_profiles" DROP COLUMN "user_id", ADD COLUMN "email" character varying NULL, ADD COLUMN "password_hash" character varying NULL, ADD COLUMN "naver_id" character varying NULL, ADD COLUMN "auth_provider" character varying NOT NULL DEFAULT 'email', ADD COLUMN "role" character varying NOT NULL DEFAULT 'user', ADD COLUMN "email_verified" boolean NOT NULL DEFAULT false, ADD COLUMN "last_login_at" timestamptz NULL;
-- Create index "userprofile_email" to table: "user_profiles"
CREATE UNIQUE INDEX "userprofile_email" ON "public"."user_profiles" ("email");
-- Create index "userprofile_naver_id" to table: "user_profiles"
CREATE UNIQUE INDEX "userprofile_naver_id" ON "public"."user_profiles" ("naver_id");
-- Create index "userprofile_role" to table: "user_profiles"
CREATE INDEX "userprofile_role" ON "public"."user_profiles" ("role");

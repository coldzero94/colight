-- Modify "applications" table
ALTER TABLE "public"."applications" ALTER COLUMN "position" DROP NOT NULL, ADD COLUMN "tags" jsonb NULL;
-- Create index "application_user_id_company_name" to table: "applications"
CREATE INDEX "application_user_id_company_name" ON "public"."applications" ("user_id", "company_name");

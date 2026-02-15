-- Modify "company_analysis_caches" table
ALTER TABLE "public"."company_analysis_caches" ADD COLUMN "view_count" bigint NOT NULL DEFAULT 0;
-- Modify "cover_letter_versions" table
ALTER TABLE "public"."cover_letter_versions" ADD COLUMN "feedback" jsonb NULL;
-- Modify "experiences" table
ALTER TABLE "public"."experiences" DROP COLUMN "embedding";
-- Modify "user_profiles" table
ALTER TABLE "public"."user_profiles" ADD COLUMN "plan" character varying NOT NULL DEFAULT 'free';
-- Create "feedbacks" table
CREATE TABLE "public"."feedbacks" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "category" character varying NOT NULL,
  "content" text NOT NULL,
  "page_url" character varying NULL,
  "user_agent" character varying NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "feedbacks_user_profiles_feedbacks" FOREIGN KEY ("user_id") REFERENCES "public"."user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "feedback_user_id_created_at" to table: "feedbacks"
CREATE INDEX "feedback_user_id_created_at" ON "public"."feedbacks" ("user_id", "created_at");
-- Create "usage_logs" table
CREATE TABLE "public"."usage_logs" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "feature" character varying NOT NULL,
  "metadata" jsonb NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "usage_logs_user_profiles_usage_logs" FOREIGN KEY ("user_id") REFERENCES "public"."user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "usagelog_user_id_feature_created_at" to table: "usage_logs"
CREATE INDEX "usagelog_user_id_feature_created_at" ON "public"."usage_logs" ("user_id", "feature", "created_at");

-- Modify "feedbacks" table
ALTER TABLE "public"."feedbacks" ADD COLUMN "admin_status" character varying NOT NULL DEFAULT 'pending', ADD COLUMN "admin_note" text NULL, ADD COLUMN "reviewed_by" uuid NULL, ADD COLUMN "reviewed_at" timestamptz NULL;
-- Modify "user_profiles" table
ALTER TABLE "public"."user_profiles" ADD COLUMN "force_logout_at" timestamptz NULL;
-- Create "deletion_requests" table
CREATE TABLE "public"."deletion_requests" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "reason" text NULL,
  "scheduled_at" timestamptz NOT NULL,
  "status" character varying NOT NULL DEFAULT 'pending',
  "requested_by" uuid NOT NULL,
  "cancelled_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "deletion_requests_user_profiles_deletion_requests" FOREIGN KEY ("user_id") REFERENCES "public"."user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "deletionrequest_status" to table: "deletion_requests"
CREATE INDEX "deletionrequest_status" ON "public"."deletion_requests" ("status");
-- Create index "deletionrequest_user_id" to table: "deletion_requests"
CREATE INDEX "deletionrequest_user_id" ON "public"."deletion_requests" ("user_id");

-- Phase 1.7: AI observability — ai_call_errors + quota_hit_events
-- Migrates error_message from usage_logs into ai_call_errors, then drops the column.

-- Create "ai_call_errors" table (1:1 extension of usage_logs for error details)
CREATE TABLE "public"."ai_call_errors" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "usage_log_id" uuid NOT NULL,
  "error_type" character varying NOT NULL,
  "error_message" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ai_call_errors_usage_logs_error_detail" UNIQUE ("usage_log_id"),
  CONSTRAINT "ai_call_errors_usage_logs_error_detail_fk" FOREIGN KEY ("usage_log_id")
    REFERENCES "public"."usage_logs" ("id") ON DELETE CASCADE
);
-- Create index "aicallerror_error_type_created_at" to table: "ai_call_errors"
CREATE INDEX "aicallerror_error_type_created_at" ON "public"."ai_call_errors" ("error_type", "created_at");
-- Create index "aicallerror_created_at" to table: "ai_call_errors"
CREATE INDEX "aicallerror_created_at" ON "public"."ai_call_errors" ("created_at");

-- Migrate existing error data from usage_logs → ai_call_errors (error_type='unknown')
INSERT INTO "public"."ai_call_errors" ("id", "created_at", "usage_log_id", "error_type", "error_message")
SELECT gen_random_uuid(), created_at, id, 'unknown', error_message
FROM "public"."usage_logs"
WHERE status = 'error' AND error_message IS NOT NULL;

-- Drop error_message column from usage_logs
ALTER TABLE "public"."usage_logs" DROP COLUMN IF EXISTS "error_message";

-- Create "quota_hit_events" table (persistent rate limit event log)
CREATE TABLE "public"."quota_hit_events" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "provider" character varying NOT NULL,
  "model" character varying NOT NULL,
  "feature" character varying NULL,
  "error_message" text NOT NULL,
  "usage_log_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "quota_hit_events_usage_logs_usage_log" FOREIGN KEY ("usage_log_id")
    REFERENCES "public"."usage_logs" ("id") ON DELETE SET NULL
);
-- Create index "quotahitevent_provider_model_created_at" to table: "quota_hit_events"
CREATE INDEX "quotahitevent_provider_model_created_at" ON "public"."quota_hit_events" ("provider", "model", "created_at");
-- Create index "quotahitevent_created_at" to table: "quota_hit_events"
CREATE INDEX "quotahitevent_created_at" ON "public"."quota_hit_events" ("created_at");

-- Create "admin_audit_logs" table
CREATE TABLE "public"."admin_audit_logs" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "admin_id" uuid NOT NULL,
  "action" character varying NOT NULL,
  "target_type" character varying NOT NULL,
  "target_id" character varying NULL,
  "old_value" text NULL,
  "new_value" text NULL,
  "ip_address" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create index "adminauditlog_action" to table: "admin_audit_logs"
CREATE INDEX "adminauditlog_action" ON "public"."admin_audit_logs" ("action");
-- Create index "adminauditlog_admin_id" to table: "admin_audit_logs"
CREATE INDEX "adminauditlog_admin_id" ON "public"."admin_audit_logs" ("admin_id");
-- Create index "adminauditlog_created_at" to table: "admin_audit_logs"
CREATE INDEX "adminauditlog_created_at" ON "public"."admin_audit_logs" ("created_at");

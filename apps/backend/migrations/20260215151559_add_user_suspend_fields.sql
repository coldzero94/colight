-- Modify "user_profiles" table
ALTER TABLE "public"."user_profiles" ADD COLUMN "suspended" boolean NOT NULL DEFAULT false, ADD COLUMN "suspended_at" timestamptz NULL, ADD COLUMN "suspended_reason" character varying NULL;

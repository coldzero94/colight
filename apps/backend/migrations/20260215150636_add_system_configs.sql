-- Create "system_configs" table
CREATE TABLE "public"."system_configs" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "config_key" character varying NOT NULL,
  "config_value" text NOT NULL,
  "description" character varying NULL,
  "category" character varying NOT NULL,
  "is_secret" boolean NOT NULL DEFAULT false,
  "updated_by" uuid NULL,
  PRIMARY KEY ("id")
);
-- Create index "system_configs_config_key_key" to table: "system_configs"
CREATE UNIQUE INDEX "system_configs_config_key_key" ON "public"."system_configs" ("config_key");
-- Create index "systemconfig_category" to table: "system_configs"
CREATE INDEX "systemconfig_category" ON "public"."system_configs" ("category");
-- Create index "systemconfig_config_key" to table: "system_configs"
CREATE UNIQUE INDEX "systemconfig_config_key" ON "public"."system_configs" ("config_key");

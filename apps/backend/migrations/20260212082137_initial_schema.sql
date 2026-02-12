-- Create pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create "company_analysis_caches" table
CREATE TABLE "company_analysis_caches" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "cache_key" character varying NOT NULL,
  "cache_type" character varying NOT NULL,
  "data" jsonb NOT NULL,
  "source_url" text NULL,
  "company_name" character varying NULL,
  "expires_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "company_analysis_caches_cache_key_key" to table: "company_analysis_caches"
CREATE UNIQUE INDEX "company_analysis_caches_cache_key_key" ON "company_analysis_caches" ("cache_key");
-- Create index "companyanalysiscache_cache_key" to table: "company_analysis_caches"
CREATE INDEX "companyanalysiscache_cache_key" ON "company_analysis_caches" ("cache_key");
-- Create index "companyanalysiscache_company_name" to table: "company_analysis_caches"
CREATE INDEX "companyanalysiscache_company_name" ON "company_analysis_caches" ("company_name");
-- Create index "companyanalysiscache_expires_at" to table: "company_analysis_caches"
CREATE INDEX "companyanalysiscache_expires_at" ON "company_analysis_caches" ("expires_at");
-- Create "user_profiles" table
CREATE TABLE "user_profiles" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_id" uuid NOT NULL,
  "nickname" character varying NULL,
  "target_job" character varying NULL,
  "target_industry" character varying NULL,
  "education_level" character varying NULL,
  "graduation_year" bigint NULL,
  "experience_years" bigint NOT NULL DEFAULT 0,
  "onboarding_completed" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "user_profiles_user_id_key" to table: "user_profiles"
CREATE UNIQUE INDEX "user_profiles_user_id_key" ON "user_profiles" ("user_id");
-- Create index "userprofile_user_id" to table: "user_profiles"
CREATE INDEX "userprofile_user_id" ON "user_profiles" ("user_id");
-- Create "talent_profiles" table
CREATE TABLE "talent_profiles" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "company_name" character varying NOT NULL,
  "industry" character varying NULL,
  "core_values" jsonb NULL,
  "talent_traits" jsonb NULL,
  "culture_keywords" jsonb NULL,
  "source" text NULL,
  "verified" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "talentprofile_company_name" to table: "talent_profiles"
CREATE INDEX "talentprofile_company_name" ON "talent_profiles" ("company_name");
-- Create "company_analyses" table
CREATE TABLE "company_analyses" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "company_name" character varying NOT NULL,
  "job_url" text NULL,
  "job_posting" jsonb NULL,
  "company_info" jsonb NULL,
  "analysis_result" jsonb NULL,
  "matching_result" jsonb NULL,
  "overall_fit_score" bigint NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "company_analyses_user_profiles_company_analyses" FOREIGN KEY ("user_id") REFERENCES "user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "companyanalysis_user_id" to table: "company_analyses"
CREATE INDEX "companyanalysis_user_id" ON "company_analyses" ("user_id");
-- Create index "companyanalysis_user_id_company_name" to table: "company_analyses"
CREATE INDEX "companyanalysis_user_id_company_name" ON "company_analyses" ("user_id", "company_name");
-- Create "applications" table
CREATE TABLE "applications" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "company_name" character varying NOT NULL,
  "position" character varying NOT NULL,
  "job_url" text NULL,
  "status" character varying NOT NULL DEFAULT 'preparing',
  "deadline" timestamptz NULL,
  "applied_at" timestamptz NULL,
  "notes" text NULL,
  "company_analysis_applications" uuid NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "applications_company_analyses_applications" FOREIGN KEY ("company_analysis_applications") REFERENCES "company_analyses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "applications_user_profiles_applications" FOREIGN KEY ("user_id") REFERENCES "user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "application_user_id" to table: "applications"
CREATE INDEX "application_user_id" ON "applications" ("user_id");
-- Create index "application_user_id_deadline" to table: "applications"
CREATE INDEX "application_user_id_deadline" ON "applications" ("user_id", "deadline");
-- Create index "application_user_id_status" to table: "applications"
CREATE INDEX "application_user_id_status" ON "applications" ("user_id", "status");
-- Create "cover_letters" table
CREATE TABLE "cover_letters" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "company_name" character varying NULL,
  "question_text" text NOT NULL,
  "question_analysis" jsonb NULL,
  "char_limit" bigint NULL,
  "current_content" text NULL,
  "status" character varying NOT NULL DEFAULT 'draft',
  "application_cover_letters" uuid NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "cover_letters_applications_cover_letters" FOREIGN KEY ("application_cover_letters") REFERENCES "applications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "cover_letters_user_profiles_cover_letters" FOREIGN KEY ("user_id") REFERENCES "user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "coverletter_user_id" to table: "cover_letters"
CREATE INDEX "coverletter_user_id" ON "cover_letters" ("user_id");
-- Create index "coverletter_user_id_status" to table: "cover_letters"
CREATE INDEX "coverletter_user_id_status" ON "cover_letters" ("user_id", "status");
-- Create "prompt_templates" table
CREATE TABLE "prompt_templates" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "category" character varying NOT NULL,
  "sub_category" character varying NOT NULL,
  "name" character varying NOT NULL,
  "system_prompt" text NOT NULL,
  "user_prompt_template" text NOT NULL,
  "output_schema" jsonb NULL,
  "model" character varying NOT NULL,
  "temperature" double precision NOT NULL DEFAULT 0.3,
  "max_tokens" bigint NOT NULL DEFAULT 2000,
  "version" bigint NOT NULL DEFAULT 1,
  "is_active" boolean NOT NULL DEFAULT true,
  "usage_count" bigint NOT NULL DEFAULT 0,
  "avg_latency_ms" bigint NOT NULL DEFAULT 0,
  "avg_quality_score" double precision NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create index "prompttemplate_category_sub_category" to table: "prompt_templates"
CREATE INDEX "prompttemplate_category_sub_category" ON "prompt_templates" ("category", "sub_category");
-- Create index "prompttemplate_is_active_category" to table: "prompt_templates"
CREATE INDEX "prompttemplate_is_active_category" ON "prompt_templates" ("is_active", "category");
-- Create "coaching_sessions" table
CREATE TABLE "coaching_sessions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "session_type" character varying NOT NULL,
  "input_data" jsonb NULL,
  "output_data" jsonb NULL,
  "model_used" character varying NULL,
  "input_tokens" bigint NOT NULL DEFAULT 0,
  "output_tokens" bigint NOT NULL DEFAULT 0,
  "total_cost_krw" double precision NOT NULL DEFAULT 0,
  "latency_ms" bigint NOT NULL DEFAULT 0,
  "quality_score" double precision NULL,
  "cover_letter_coaching_sessions" uuid NULL,
  "prompt_template_coaching_sessions" uuid NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "coaching_sessions_cover_letters_coaching_sessions" FOREIGN KEY ("cover_letter_coaching_sessions") REFERENCES "cover_letters" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "coaching_sessions_prompt_templates_coaching_sessions" FOREIGN KEY ("prompt_template_coaching_sessions") REFERENCES "prompt_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "coaching_sessions_user_profiles_coaching_sessions" FOREIGN KEY ("user_id") REFERENCES "user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "coachingsession_session_type" to table: "coaching_sessions"
CREATE INDEX "coachingsession_session_type" ON "coaching_sessions" ("session_type");
-- Create index "coachingsession_user_id" to table: "coaching_sessions"
CREATE INDEX "coachingsession_user_id" ON "coaching_sessions" ("user_id");
-- Create "cover_letter_versions" table
CREATE TABLE "cover_letter_versions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "version_number" bigint NOT NULL,
  "content" text NOT NULL,
  "char_count" bigint NULL,
  "change_summary" text NULL,
  "scores" jsonb NULL,
  "coaching_session_cover_letter_versions" uuid NULL,
  "cover_letter_versions" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "cover_letter_versions_coaching_sessions_cover_letter_versions" FOREIGN KEY ("coaching_session_cover_letter_versions") REFERENCES "coaching_sessions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "cover_letter_versions_cover_letters_versions" FOREIGN KEY ("cover_letter_versions") REFERENCES "cover_letters" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "coverletterversion_version_number" to table: "cover_letter_versions"
CREATE INDEX "coverletterversion_version_number" ON "cover_letter_versions" ("version_number");
-- Create "experiences" table
CREATE TABLE "experiences" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "title" character varying NOT NULL,
  "category" character varying NULL,
  "period_start" timestamptz NULL,
  "period_end" timestamptz NULL,
  "role" character varying NULL,
  "content" text NOT NULL,
  "result" text NULL,
  "star_situation" text NULL,
  "star_task" text NULL,
  "star_action" text NULL,
  "star_result" text NULL,
  "keywords" jsonb NULL,
  "source" character varying NOT NULL DEFAULT 'manual',
  "is_archived" boolean NOT NULL DEFAULT false,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "experiences_user_profiles_experiences" FOREIGN KEY ("user_id") REFERENCES "user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "experience_user_id" to table: "experiences"
CREATE INDEX "experience_user_id" ON "experiences" ("user_id");
-- Create index "experience_user_id_category" to table: "experiences"
CREATE INDEX "experience_user_id_category" ON "experiences" ("user_id", "category");
-- Create index "experience_user_id_created_at" to table: "experiences"
CREATE INDEX "experience_user_id_created_at" ON "experiences" ("user_id", "created_at");
-- Create index "experience_user_id_is_archived" to table: "experiences"
CREATE INDEX "experience_user_id_is_archived" ON "experiences" ("user_id", "is_archived");
-- Create "experience_tags" table
CREATE TABLE "experience_tags" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "tag_name" character varying NOT NULL,
  "tag_type" character varying NOT NULL,
  "confidence" double precision NOT NULL DEFAULT 0,
  "experience_tags" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "experience_tags_experiences_tags" FOREIGN KEY ("experience_tags") REFERENCES "experiences" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "experiencetag_tag_name" to table: "experience_tags"
CREATE INDEX "experiencetag_tag_name" ON "experience_tags" ("tag_name");
-- Create "experience_usages" table
CREATE TABLE "experience_usages" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "question_text" text NULL,
  "company_name" character varying NULL,
  "used_at" timestamptz NOT NULL,
  "application_experience_usages" uuid NULL,
  "cover_letter_experience_usages" uuid NULL,
  "experience_usages" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "experience_usages_applications_experience_usages" FOREIGN KEY ("application_experience_usages") REFERENCES "applications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "experience_usages_cover_letters_experience_usages" FOREIGN KEY ("cover_letter_experience_usages") REFERENCES "cover_letters" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "experience_usages_experiences_usages" FOREIGN KEY ("experience_usages") REFERENCES "experiences" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "experience_usages_user_profiles_experience_usages" FOREIGN KEY ("user_id") REFERENCES "user_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "experienceusage_user_id" to table: "experience_usages"
CREATE INDEX "experienceusage_user_id" ON "experience_usages" ("user_id");
-- Create "weapon_categories" table
CREATE TABLE "weapon_categories" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "code" character varying NOT NULL,
  "parent_code" character varying NULL,
  "name" character varying NOT NULL,
  "description" text NULL,
  "keywords" jsonb NULL,
  "question_patterns" jsonb NULL,
  "display_order" bigint NOT NULL DEFAULT 0,
  "icon" character varying NULL,
  "color" character varying NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "weapon_categories_code_key" to table: "weapon_categories"
CREATE UNIQUE INDEX "weapon_categories_code_key" ON "weapon_categories" ("code");
-- Create index "weaponcategory_code" to table: "weapon_categories"
CREATE INDEX "weaponcategory_code" ON "weapon_categories" ("code");
-- Create index "weaponcategory_parent_code" to table: "weapon_categories"
CREATE INDEX "weaponcategory_parent_code" ON "weapon_categories" ("parent_code");
-- Create "experience_weapons" table
CREATE TABLE "experience_weapons" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "weapon_code" character varying NOT NULL,
  "confidence" double precision NOT NULL DEFAULT 0,
  "is_primary" boolean NOT NULL DEFAULT false,
  "reasoning" text NULL,
  "user_confirmed" boolean NOT NULL DEFAULT false,
  "user_modified" boolean NOT NULL DEFAULT false,
  "experience_weapons" uuid NOT NULL,
  "weapon_category_experience_weapons" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "experience_weapons_experiences_weapons" FOREIGN KEY ("experience_weapons") REFERENCES "experiences" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "experience_weapons_weapon_categories_experience_weapons" FOREIGN KEY ("weapon_category_experience_weapons") REFERENCES "weapon_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "experienceweapon_is_primary" to table: "experience_weapons"
CREATE INDEX "experienceweapon_is_primary" ON "experience_weapons" ("is_primary");
-- Create index "experienceweapon_weapon_code" to table: "experience_weapons"
CREATE INDEX "experienceweapon_weapon_code" ON "experience_weapons" ("weapon_code");
-- Create "question_patterns" table
CREATE TABLE "question_patterns" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "pattern_type" character varying NOT NULL,
  "pattern_name" character varying NOT NULL,
  "detection_keywords" jsonb NULL,
  "detection_regex" text NULL,
  "primary_weapons" jsonb NULL,
  "secondary_weapons" jsonb NULL,
  "writing_guide" jsonb NULL,
  "display_order" bigint NOT NULL DEFAULT 0,
  "is_active" boolean NOT NULL DEFAULT true,
  "prompt_template_question_patterns" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "question_patterns_prompt_templates_question_patterns" FOREIGN KEY ("prompt_template_question_patterns") REFERENCES "prompt_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "questionpattern_is_active" to table: "question_patterns"
CREATE INDEX "questionpattern_is_active" ON "question_patterns" ("is_active");
-- Create index "questionpattern_pattern_type" to table: "question_patterns"
CREATE INDEX "questionpattern_pattern_type" ON "question_patterns" ("pattern_type");

-- Add pgvector embedding column to experiences table
ALTER TABLE "experiences" ADD COLUMN embedding vector(1536);

-- Create vector index for cosine similarity search
CREATE INDEX idx_experiences_embedding ON "experiences" USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

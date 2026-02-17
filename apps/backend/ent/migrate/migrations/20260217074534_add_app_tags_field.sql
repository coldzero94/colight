-- Create "admin_audit_logs" table
CREATE TABLE `admin_audit_logs` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `admin_id` uuid NOT NULL,
  `action` text NOT NULL,
  `target_type` text NOT NULL,
  `target_id` text NULL,
  `old_value` text NULL,
  `new_value` text NULL,
  `ip_address` text NULL,
  PRIMARY KEY (`id`)
);
-- Create index "adminauditlog_admin_id" to table: "admin_audit_logs"
CREATE INDEX `adminauditlog_admin_id` ON `admin_audit_logs` (`admin_id`);
-- Create index "adminauditlog_action" to table: "admin_audit_logs"
CREATE INDEX `adminauditlog_action` ON `admin_audit_logs` (`action`);
-- Create index "adminauditlog_created_at" to table: "admin_audit_logs"
CREATE INDEX `adminauditlog_created_at` ON `admin_audit_logs` (`created_at`);
-- Create "applications" table
CREATE TABLE `applications` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `company_name` text NOT NULL,
  `position` text NULL,
  `job_url` text NULL,
  `status` text NOT NULL DEFAULT 'preparing',
  `deadline` datetime NULL,
  `applied_at` datetime NULL,
  `notes` text NULL,
  `tags` json NULL,
  `company_analysis_applications` uuid NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `applications_user_profiles_applications` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `applications_company_analyses_applications` FOREIGN KEY (`company_analysis_applications`) REFERENCES `company_analyses` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "application_user_id" to table: "applications"
CREATE INDEX `application_user_id` ON `applications` (`user_id`);
-- Create index "application_user_id_status" to table: "applications"
CREATE INDEX `application_user_id_status` ON `applications` (`user_id`, `status`);
-- Create index "application_user_id_deadline" to table: "applications"
CREATE INDEX `application_user_id_deadline` ON `applications` (`user_id`, `deadline`);
-- Create index "application_user_id_company_name" to table: "applications"
CREATE INDEX `application_user_id_company_name` ON `applications` (`user_id`, `company_name`);
-- Create "coaching_sessions" table
CREATE TABLE `coaching_sessions` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `session_type` text NOT NULL,
  `input_data` json NULL,
  `output_data` json NULL,
  `model_used` text NULL,
  `input_tokens` integer NOT NULL DEFAULT 0,
  `output_tokens` integer NOT NULL DEFAULT 0,
  `total_cost_krw` real NOT NULL DEFAULT 0,
  `latency_ms` integer NOT NULL DEFAULT 0,
  `quality_score` real NULL,
  `cover_letter_coaching_sessions` uuid NULL,
  `prompt_template_coaching_sessions` uuid NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `coaching_sessions_user_profiles_coaching_sessions` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `coaching_sessions_prompt_templates_coaching_sessions` FOREIGN KEY (`prompt_template_coaching_sessions`) REFERENCES `prompt_templates` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `coaching_sessions_cover_letters_coaching_sessions` FOREIGN KEY (`cover_letter_coaching_sessions`) REFERENCES `cover_letters` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "coachingsession_user_id" to table: "coaching_sessions"
CREATE INDEX `coachingsession_user_id` ON `coaching_sessions` (`user_id`);
-- Create index "coachingsession_session_type" to table: "coaching_sessions"
CREATE INDEX `coachingsession_session_type` ON `coaching_sessions` (`session_type`);
-- Create "company_analyses" table
CREATE TABLE `company_analyses` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `company_name` text NOT NULL,
  `job_url` text NULL,
  `job_posting` json NULL,
  `company_info` json NULL,
  `analysis_result` json NULL,
  `matching_result` json NULL,
  `overall_fit_score` integer NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `company_analyses_user_profiles_company_analyses` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "companyanalysis_user_id" to table: "company_analyses"
CREATE INDEX `companyanalysis_user_id` ON `company_analyses` (`user_id`);
-- Create index "companyanalysis_user_id_company_name" to table: "company_analyses"
CREATE INDEX `companyanalysis_user_id_company_name` ON `company_analyses` (`user_id`, `company_name`);
-- Create "company_analysis_caches" table
CREATE TABLE `company_analysis_caches` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `cache_key` text NOT NULL,
  `cache_type` text NOT NULL,
  `data` json NOT NULL,
  `source_url` text NULL,
  `company_name` text NULL,
  `expires_at` datetime NOT NULL,
  `view_count` integer NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
);
-- Create index "company_analysis_caches_cache_key_key" to table: "company_analysis_caches"
CREATE UNIQUE INDEX `company_analysis_caches_cache_key_key` ON `company_analysis_caches` (`cache_key`);
-- Create index "companyanalysiscache_cache_key" to table: "company_analysis_caches"
CREATE INDEX `companyanalysiscache_cache_key` ON `company_analysis_caches` (`cache_key`);
-- Create index "companyanalysiscache_expires_at" to table: "company_analysis_caches"
CREATE INDEX `companyanalysiscache_expires_at` ON `company_analysis_caches` (`expires_at`);
-- Create index "companyanalysiscache_company_name" to table: "company_analysis_caches"
CREATE INDEX `companyanalysiscache_company_name` ON `company_analysis_caches` (`company_name`);
-- Create "cover_letters" table
CREATE TABLE `cover_letters` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `company_name` text NULL,
  `question_text` text NOT NULL,
  `question_analysis` json NULL,
  `char_limit` integer NULL,
  `current_content` text NULL,
  `status` text NOT NULL DEFAULT 'draft',
  `application_cover_letters` uuid NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `cover_letters_user_profiles_cover_letters` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `cover_letters_applications_cover_letters` FOREIGN KEY (`application_cover_letters`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "coverletter_user_id" to table: "cover_letters"
CREATE INDEX `coverletter_user_id` ON `cover_letters` (`user_id`);
-- Create index "coverletter_user_id_status" to table: "cover_letters"
CREATE INDEX `coverletter_user_id_status` ON `cover_letters` (`user_id`, `status`);
-- Create "cover_letter_versions" table
CREATE TABLE `cover_letter_versions` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `version_number` integer NOT NULL,
  `content` text NOT NULL,
  `char_count` integer NULL,
  `change_summary` text NULL,
  `scores` json NULL,
  `feedback` json NULL,
  `coaching_session_cover_letter_versions` uuid NULL,
  `cover_letter_versions` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `cover_letter_versions_cover_letters_versions` FOREIGN KEY (`cover_letter_versions`) REFERENCES `cover_letters` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `cover_letter_versions_coaching_sessions_cover_letter_versions` FOREIGN KEY (`coaching_session_cover_letter_versions`) REFERENCES `coaching_sessions` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "coverletterversion_version_number" to table: "cover_letter_versions"
CREATE INDEX `coverletterversion_version_number` ON `cover_letter_versions` (`version_number`);
-- Create "deletion_requests" table
CREATE TABLE `deletion_requests` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `reason` text NULL,
  `scheduled_at` datetime NOT NULL,
  `status` text NOT NULL DEFAULT 'pending',
  `requested_by` uuid NOT NULL,
  `cancelled_at` datetime NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `deletion_requests_user_profiles_deletion_requests` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "deletionrequest_user_id" to table: "deletion_requests"
CREATE INDEX `deletionrequest_user_id` ON `deletion_requests` (`user_id`);
-- Create index "deletionrequest_status" to table: "deletion_requests"
CREATE INDEX `deletionrequest_status` ON `deletion_requests` (`status`);
-- Create "experiences" table
CREATE TABLE `experiences` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `title` text NOT NULL,
  `category` text NULL,
  `period_start` datetime NULL,
  `period_end` datetime NULL,
  `role` text NULL,
  `content` text NOT NULL,
  `result` text NULL,
  `star_situation` text NULL,
  `star_task` text NULL,
  `star_action` text NULL,
  `star_result` text NULL,
  `keywords` json NULL,
  `source` text NOT NULL DEFAULT 'manual',
  `is_archived` bool NOT NULL DEFAULT false,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `experiences_user_profiles_experiences` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "experience_user_id" to table: "experiences"
CREATE INDEX `experience_user_id` ON `experiences` (`user_id`);
-- Create index "experience_user_id_category" to table: "experiences"
CREATE INDEX `experience_user_id_category` ON `experiences` (`user_id`, `category`);
-- Create index "experience_user_id_created_at" to table: "experiences"
CREATE INDEX `experience_user_id_created_at` ON `experiences` (`user_id`, `created_at`);
-- Create index "experience_user_id_is_archived" to table: "experiences"
CREATE INDEX `experience_user_id_is_archived` ON `experiences` (`user_id`, `is_archived`);
-- Create "experience_tags" table
CREATE TABLE `experience_tags` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `tag_name` text NOT NULL,
  `tag_type` text NOT NULL,
  `confidence` real NOT NULL DEFAULT 0,
  `experience_tags` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `experience_tags_experiences_tags` FOREIGN KEY (`experience_tags`) REFERENCES `experiences` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "experiencetag_tag_name" to table: "experience_tags"
CREATE INDEX `experiencetag_tag_name` ON `experience_tags` (`tag_name`);
-- Create "experience_usages" table
CREATE TABLE `experience_usages` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `question_text` text NULL,
  `company_name` text NULL,
  `used_at` datetime NOT NULL,
  `application_experience_usages` uuid NULL,
  `cover_letter_experience_usages` uuid NULL,
  `experience_usages` uuid NOT NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `experience_usages_user_profiles_experience_usages` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `experience_usages_experiences_usages` FOREIGN KEY (`experience_usages`) REFERENCES `experiences` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `experience_usages_cover_letters_experience_usages` FOREIGN KEY (`cover_letter_experience_usages`) REFERENCES `cover_letters` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `experience_usages_applications_experience_usages` FOREIGN KEY (`application_experience_usages`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "experienceusage_user_id" to table: "experience_usages"
CREATE INDEX `experienceusage_user_id` ON `experience_usages` (`user_id`);
-- Create "experience_weapons" table
CREATE TABLE `experience_weapons` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `weapon_code` text NOT NULL,
  `confidence` real NOT NULL DEFAULT 0,
  `is_primary` bool NOT NULL DEFAULT false,
  `reasoning` text NULL,
  `user_confirmed` bool NOT NULL DEFAULT false,
  `user_modified` bool NOT NULL DEFAULT false,
  `experience_weapons` uuid NOT NULL,
  `weapon_category_experience_weapons` uuid NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `experience_weapons_weapon_categories_experience_weapons` FOREIGN KEY (`weapon_category_experience_weapons`) REFERENCES `weapon_categories` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `experience_weapons_experiences_weapons` FOREIGN KEY (`experience_weapons`) REFERENCES `experiences` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "experienceweapon_weapon_code" to table: "experience_weapons"
CREATE INDEX `experienceweapon_weapon_code` ON `experience_weapons` (`weapon_code`);
-- Create index "experienceweapon_is_primary" to table: "experience_weapons"
CREATE INDEX `experienceweapon_is_primary` ON `experience_weapons` (`is_primary`);
-- Create "feedbacks" table
CREATE TABLE `feedbacks` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `category` text NOT NULL,
  `content` text NOT NULL,
  `page_url` text NULL,
  `user_agent` text NULL,
  `admin_status` text NOT NULL DEFAULT 'pending',
  `admin_note` text NULL,
  `reviewed_by` uuid NULL,
  `reviewed_at` datetime NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `feedbacks_user_profiles_feedbacks` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "feedback_user_id_created_at" to table: "feedbacks"
CREATE INDEX `feedback_user_id_created_at` ON `feedbacks` (`user_id`, `created_at`);
-- Create "prompt_templates" table
CREATE TABLE `prompt_templates` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `category` text NOT NULL,
  `sub_category` text NOT NULL,
  `name` text NOT NULL,
  `system_prompt` text NOT NULL,
  `user_prompt_template` text NOT NULL,
  `output_schema` json NULL,
  `model` text NOT NULL,
  `temperature` real NOT NULL DEFAULT 0.3,
  `max_tokens` integer NOT NULL DEFAULT 2000,
  `version` integer NOT NULL DEFAULT 1,
  `is_active` bool NOT NULL DEFAULT true,
  `usage_count` integer NOT NULL DEFAULT 0,
  `avg_latency_ms` integer NOT NULL DEFAULT 0,
  `avg_quality_score` real NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
);
-- Create index "prompttemplate_category_sub_category" to table: "prompt_templates"
CREATE INDEX `prompttemplate_category_sub_category` ON `prompt_templates` (`category`, `sub_category`);
-- Create index "prompttemplate_is_active_category" to table: "prompt_templates"
CREATE INDEX `prompttemplate_is_active_category` ON `prompt_templates` (`is_active`, `category`);
-- Create "question_patterns" table
CREATE TABLE `question_patterns` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `pattern_type` text NOT NULL,
  `pattern_name` text NOT NULL,
  `detection_keywords` json NULL,
  `detection_regex` text NULL,
  `primary_weapons` json NULL,
  `secondary_weapons` json NULL,
  `writing_guide` json NULL,
  `display_order` integer NOT NULL DEFAULT 0,
  `is_active` bool NOT NULL DEFAULT true,
  `prompt_template_question_patterns` uuid NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `question_patterns_prompt_templates_question_patterns` FOREIGN KEY (`prompt_template_question_patterns`) REFERENCES `prompt_templates` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "questionpattern_pattern_type" to table: "question_patterns"
CREATE INDEX `questionpattern_pattern_type` ON `question_patterns` (`pattern_type`);
-- Create index "questionpattern_is_active" to table: "question_patterns"
CREATE INDEX `questionpattern_is_active` ON `question_patterns` (`is_active`);
-- Create "system_configs" table
CREATE TABLE `system_configs` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `config_key` text NOT NULL,
  `config_value` text NOT NULL,
  `description` text NULL,
  `category` text NOT NULL,
  `is_secret` bool NOT NULL DEFAULT false,
  `updated_by` uuid NULL,
  PRIMARY KEY (`id`)
);
-- Create index "system_configs_config_key_key" to table: "system_configs"
CREATE UNIQUE INDEX `system_configs_config_key_key` ON `system_configs` (`config_key`);
-- Create index "systemconfig_category" to table: "system_configs"
CREATE INDEX `systemconfig_category` ON `system_configs` (`category`);
-- Create index "systemconfig_config_key" to table: "system_configs"
CREATE UNIQUE INDEX `systemconfig_config_key` ON `system_configs` (`config_key`);
-- Create "talent_profiles" table
CREATE TABLE `talent_profiles` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `company_name` text NOT NULL,
  `industry` text NULL,
  `core_values` json NULL,
  `talent_traits` json NULL,
  `culture_keywords` json NULL,
  `source` text NULL,
  `verified` bool NOT NULL DEFAULT false,
  PRIMARY KEY (`id`)
);
-- Create index "talentprofile_company_name" to table: "talent_profiles"
CREATE INDEX `talentprofile_company_name` ON `talent_profiles` (`company_name`);
-- Create "usage_logs" table
CREATE TABLE `usage_logs` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `feature` text NOT NULL,
  `metadata` json NULL,
  `provider` text NULL,
  `model` text NULL,
  `input_tokens` integer NOT NULL DEFAULT 0,
  `output_tokens` integer NOT NULL DEFAULT 0,
  `total_tokens` integer NOT NULL DEFAULT 0,
  `estimated_cost_krw` real NULL,
  `latency_ms` integer NULL,
  `status` text NOT NULL DEFAULT 'success',
  `error_message` text NULL,
  `user_id` uuid NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `usage_logs_user_profiles_usage_logs` FOREIGN KEY (`user_id`) REFERENCES `user_profiles` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "usagelog_user_id_feature_created_at" to table: "usage_logs"
CREATE INDEX `usagelog_user_id_feature_created_at` ON `usage_logs` (`user_id`, `feature`, `created_at`);
-- Create "user_profiles" table
CREATE TABLE `user_profiles` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `email` text NULL,
  `password_hash` text NULL,
  `naver_id` text NULL,
  `auth_provider` text NOT NULL DEFAULT 'email',
  `role` text NOT NULL DEFAULT 'user',
  `email_verified` bool NOT NULL DEFAULT false,
  `last_login_at` datetime NULL,
  `nickname` text NULL,
  `target_job` text NULL,
  `target_industry` text NULL,
  `education_level` text NULL,
  `graduation_year` integer NULL,
  `experience_years` integer NOT NULL DEFAULT 0,
  `onboarding_completed` bool NOT NULL DEFAULT false,
  `plan` text NOT NULL DEFAULT 'free',
  `suspended` bool NOT NULL DEFAULT false,
  `suspended_at` datetime NULL,
  `suspended_reason` text NULL,
  `force_logout_at` datetime NULL,
  PRIMARY KEY (`id`)
);
-- Create index "userprofile_email" to table: "user_profiles"
CREATE UNIQUE INDEX `userprofile_email` ON `user_profiles` (`email`);
-- Create index "userprofile_naver_id" to table: "user_profiles"
CREATE UNIQUE INDEX `userprofile_naver_id` ON `user_profiles` (`naver_id`);
-- Create index "userprofile_role" to table: "user_profiles"
CREATE INDEX `userprofile_role` ON `user_profiles` (`role`);
-- Create "weapon_categories" table
CREATE TABLE `weapon_categories` (
  `id` uuid NOT NULL,
  `created_at` datetime NOT NULL,
  `code` text NOT NULL,
  `parent_code` text NULL,
  `name` text NOT NULL,
  `description` text NULL,
  `keywords` json NULL,
  `question_patterns` json NULL,
  `display_order` integer NOT NULL DEFAULT 0,
  `icon` text NULL,
  `color` text NULL,
  `is_active` bool NOT NULL DEFAULT true,
  PRIMARY KEY (`id`)
);
-- Create index "weapon_categories_code_key" to table: "weapon_categories"
CREATE UNIQUE INDEX `weapon_categories_code_key` ON `weapon_categories` (`code`);
-- Create index "weaponcategory_parent_code" to table: "weapon_categories"
CREATE INDEX `weaponcategory_parent_code` ON `weapon_categories` (`parent_code`);
-- Create index "weaponcategory_code" to table: "weapon_categories"
CREATE INDEX `weaponcategory_code` ON `weapon_categories` (`code`);

import "@typespec/http";
import "../common/errors.tsp";
import "../auth/auth.tsp";

using TypeSpec.Http;

namespace Colight.Admin;

// ─── Models ───

model AdminUserListItem {
  id: string;
  email?: string;
  nickname?: string;
  role: "user" | "admin";

  @encodedName("application/json", "auth_provider")
  authProvider: "email" | "naver";

  @encodedName("application/json", "experience_count")
  experienceCount: int32;

  @encodedName("application/json", "last_login_at")
  lastLoginAt?: utcDateTime;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model UpdateRoleRequest {
  role: "user" | "admin";
}

model SystemStats {
  @encodedName("application/json", "total_users")
  totalUsers: int32;

  @encodedName("application/json", "active_users_today")
  activeUsersToday: int32;

  @encodedName("application/json", "total_experiences")
  totalExperiences: int32;

  @encodedName("application/json", "total_coaching_sessions")
  totalCoachingSessions: int32;

  @encodedName("application/json", "total_analyses")
  totalAnalyses: int32;

  @encodedName("application/json", "naver_auth_users")
  naverAuthUsers: int32;

  @encodedName("application/json", "email_auth_users")
  emailAuthUsers: int32;
}

model PromptTemplateDetail {
  id: string;
  category: string;

  @encodedName("application/json", "sub_category")
  subCategory: string;

  name: string;

  @encodedName("application/json", "system_prompt")
  systemPrompt: string;

  @encodedName("application/json", "user_prompt_template")
  userPromptTemplate: string;

  @encodedName("application/json", "model_name")
  modelName: string;

  temperature: float32;

  @encodedName("application/json", "max_tokens")
  maxTokens: int32;

  version: int32;

  @encodedName("application/json", "is_active")
  isActive: boolean;

  @encodedName("application/json", "usage_count")
  usageCount: int32;
}

model UpdatePromptRequest {
  @encodedName("application/json", "system_prompt")
  systemPrompt?: string;

  @encodedName("application/json", "user_prompt_template")
  userPromptTemplate?: string;

  temperature?: float32;

  @encodedName("application/json", "max_tokens")
  maxTokens?: int32;

  @encodedName("application/json", "is_active")
  isActive?: boolean;
}

// ─── Admin Interface ───

@route("/v1/admin")
@tag("Admin")
@useAuth(BearerAuth)
interface AdminAPI {
  /** List users with pagination and filters */
  @get
  @route("/users")
  listUsers(
    @query limit?: int32 = 20,
    @query offset?: int32 = 0,
    @query role?: "user" | "admin",
    @query search?: string,
  ): {
    @statusCode statusCode: 200;
    @body body: { data: AdminUserListItem[]; count: int32 };
  } | Colight.Common.ErrorResponse;

  /** Get user details */
  @get
  @route("/users/{id}")
  getUser(@path id: string): {
    @statusCode statusCode: 200;
    @body body: { data: AdminUserListItem };
  } | Colight.Common.ErrorResponse;

  /** Update user role */
  @put
  @route("/users/{id}/role")
  updateUserRole(@path id: string, @body body: UpdateRoleRequest): {
    @statusCode statusCode: 200;
    @body body: { data: AdminUserListItem };
  } | Colight.Common.ErrorResponse;

  /** Get system statistics */
  @get
  @route("/stats")
  getStats(): {
    @statusCode statusCode: 200;
    @body body: { data: SystemStats };
  } | Colight.Common.ErrorResponse;

  /** List prompt templates */
  @get
  @route("/prompts")
  listPrompts(@query category?: string): {
    @statusCode statusCode: 200;
    @body body: { data: PromptTemplateDetail[] };
  } | Colight.Common.ErrorResponse;

  /** Update prompt template */
  @put
  @route("/prompts/{id}")
  updatePrompt(@path id: string, @body body: UpdatePromptRequest): {
    @statusCode statusCode: 200;
    @body body: { data: PromptTemplateDetail };
  } | Colight.Common.ErrorResponse;
}

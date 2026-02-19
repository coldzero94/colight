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

// ─── Phase 1.7 Models ───

model UserStatsSummary {
  @encodedName("application/json", "total_users")
  totalUsers: int32;

  @encodedName("application/json", "new_users_today")
  newUsersToday: int32;

  @encodedName("application/json", "active_users_today")
  activeUsersToday: int32;
}

model AIMetricsToday {
  calls: int32;

  @encodedName("application/json", "calls_delta_pct")
  callsDeltaPct: float32;

  @encodedName("application/json", "error_rate")
  errorRate: float32;

  @encodedName("application/json", "error_rate_delta")
  errorRateDelta: float32;

  @encodedName("application/json", "cost_krw")
  costKrw: float32;
}

model DailyMetric {
  date: string;
  calls: int32;

  @encodedName("application/json", "error_count")
  errorCount: int32;

  @encodedName("application/json", "cost_krw")
  costKrw: float32;
}

model QuotaAlert {
  @encodedName("application/json", "model")
  modelName: string;

  provider: string;

  @encodedName("application/json", "used_pct")
  usedPct: float32;

  remaining: int32;
}

model AIErrorItem {
  id: string;

  @encodedName("application/json", "error_type")
  errorType: "rate_limit" | "timeout" | "provider_error" | "invalid_request" | "context_exceeded";

  @encodedName("application/json", "error_message")
  errorMessage: string;

  @encodedName("application/json", "model")
  modelName?: string;

  provider?: string;
  feature: string;

  @encodedName("application/json", "user_id")
  userId: string;

  @encodedName("application/json", "input_tokens")
  inputTokens: int32;

  @encodedName("application/json", "output_tokens")
  outputTokens: int32;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model FeedbackPreview {
  id: string;
  category: "bug" | "improvement" | "other";

  @encodedName("application/json", "content_preview")
  contentPreview: string;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model DashboardData {
  @encodedName("application/json", "user_stats")
  userStats: UserStatsSummary;

  @encodedName("application/json", "ai_metrics_today")
  aiMetricsToday: AIMetricsToday;

  @encodedName("application/json", "daily_metrics")
  dailyMetrics: DailyMetric[];

  @encodedName("application/json", "quota_alerts")
  quotaAlerts: QuotaAlert[];

  @encodedName("application/json", "recent_errors")
  recentErrors: AIErrorItem[];

  @encodedName("application/json", "recent_feedbacks")
  recentFeedbacks: FeedbackPreview[];

  @encodedName("application/json", "pending_feedback_count")
  pendingFeedbackCount: int32;
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

  /** Get aggregated dashboard data */
  @get
  @route("/dashboard")
  getDashboard(): {
    @statusCode statusCode: 200;
    @body body: { data: DashboardData };
  } | Colight.Common.ErrorResponse;

  /** List AI errors with filters */
  @get
  @route("/usage/errors")
  listErrors(
    @query error_type?: "rate_limit" | "timeout" | "provider_error" | "invalid_request" | "context_exceeded",
    @query provider?: string,
    @query feature?: string,
    @query days?: int32 = 7,
    @query limit?: int32 = 20,
    @query offset?: int32 = 0,
  ): {
    @statusCode statusCode: 200;
    @body body: { data: AIErrorItem[]; count: int32 };
  } | Colight.Common.ErrorResponse;
}

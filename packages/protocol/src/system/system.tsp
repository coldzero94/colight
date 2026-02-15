import "@typespec/http";
import "../common/errors.tsp";

using TypeSpec.Http;

namespace Colight.System;

// ─── Models ───

model FeatureUsage {
  allowed: boolean;
  used: int32;
  limit: int32;
  remaining: int32;
}

model UsageResponse {
  plan: "free" | "starter" | "pro" | "season";
  features: Record<FeatureUsage>;
}

model FeedbackInput {
  category: "bug" | "improvement" | "other";
  content: string;

  @encodedName("application/json", "page_url")
  pageUrl?: string;
}

model FeedbackResponse {
  id: string;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

// ─── System Interface ───

@tag("System")
@useAuth(BearerAuth)
interface SystemAPI {
  /** Get usage limits and current consumption */
  @get
  @route("/v1/usage")
  getUsage(): {
    @statusCode statusCode: 200;
    @body body: UsageResponse;
  } | Colight.Common.ErrorResponse;

  /** Submit user feedback */
  @post
  @route("/v1/feedback")
  submitFeedback(@body body: FeedbackInput): {
    @statusCode statusCode: 201;
    @body body: FeedbackResponse;
  } | Colight.Common.ErrorResponse;
}

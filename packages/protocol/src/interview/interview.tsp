import "@typespec/http";
import "../common/errors.tsp";

using TypeSpec.Http;

namespace Colight.Interview;

// ─── Models ───

model ChatMessage {
  role: "user" | "assistant";
  content: string;
}

model GenerateQuestionRequest {
  stage: "warmup" | "memory" | "challenge" | "solution" | "outcome";
  messages: ChatMessage[];
}

model GenerateQuestionResponse {
  question: string;
  stage: "warmup" | "memory" | "challenge" | "solution" | "outcome";

  @encodedName("application/json", "next_stage")
  nextStage: string;

  @encodedName("application/json", "is_complete")
  isComplete: boolean;
}

model ExtractSTARRequest {
  messages: ChatMessage[];
}

model ExtractSTARResult {
  title: string;
  category: string;
  content: string;
  result: string;

  @encodedName("application/json", "star_situation")
  starSituation: string;

  @encodedName("application/json", "star_task")
  starTask: string;

  @encodedName("application/json", "star_action")
  starAction: string;

  @encodedName("application/json", "star_result")
  starResult: string;

  keywords: string[];
}

model SaveExperienceResponse {
  @encodedName("application/json", "experience_id")
  experienceId: string;

  tagged: boolean;
}

// ─── Interview Interface ───

@route("/v1/interview")
@tag("Interview")
@useAuth(BearerAuth)
interface InterviewAPI {
  /** Generate next interview question based on stage and conversation */
  @post
  @route("/question")
  generateQuestion(@body body: GenerateQuestionRequest): {
    @statusCode statusCode: 200;
    @body body: GenerateQuestionResponse;
  } | Colight.Common.ErrorResponse;

  /** Extract STAR structure from interview conversation */
  @post
  @route("/extract")
  extractSTAR(@body body: ExtractSTARRequest): {
    @statusCode statusCode: 200;
    @body body: ExtractSTARResult;
  } | Colight.Common.ErrorResponse;

  /** Save extracted experience with auto weapon tagging */
  @post
  @route("/save")
  saveExperience(@body body: ExtractSTARResult): {
    @statusCode statusCode: 201;
    @body body: SaveExperienceResponse;
  } | Colight.Common.ErrorResponse;
}

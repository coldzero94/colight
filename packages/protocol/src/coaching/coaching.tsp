import "@typespec/http";
import "../common/errors.tsp";

using TypeSpec.Http;

namespace Colight.Coaching;

// ─── Models ───

model WeaponRef {
  @encodedName("application/json", "weapon_id")
  weaponId: string;

  @encodedName("application/json", "weapon_name")
  weaponName: string;

  reason: string;
}

model WritingSection {
  name: string;

  @encodedName("application/json", "char_ratio")
  charRatio: float32;

  @encodedName("application/json", "char_count")
  charCount: int32;

  guide: string;
}

model RealIntent {
  intent: string;
  why: string;
}

model QuestionAnalysisRequest {
  @encodedName("application/json", "application_id")
  applicationId: string;

  @encodedName("application/json", "question_text")
  questionText: string;

  @encodedName("application/json", "char_limit")
  charLimit: int32;
}

model QuestionAnalysisResult {
  @encodedName("application/json", "surface_question")
  surfaceQuestion: string;

  @encodedName("application/json", "real_intents")
  realIntents: RealIntent[];

  @encodedName("application/json", "required_weapons")
  requiredWeapons: {
    primary: WeaponRef;
    secondary: WeaponRef[];
  };

  @encodedName("application/json", "writing_structure")
  writingStructure: {
    @encodedName("application/json", "total_chars")
    totalChars: int32;

    sections: WritingSection[];
  };

  @encodedName("application/json", "key_keywords")
  keyKeywords: string[];

  @encodedName("application/json", "avoid_list")
  avoidList: string[];

  @encodedName("application/json", "good_structure_example")
  goodStructureExample: string;
}

model GenerateDraftRequest {
  @encodedName("application/json", "application_id")
  applicationId: string;

  @encodedName("application/json", "experience_ids")
  experienceIds: string[];

  @encodedName("application/json", "question_text")
  questionText: string;

  @encodedName("application/json", "char_limit")
  charLimit: int32;

  @encodedName("application/json", "analysis_result")
  analysisResult?: QuestionAnalysisResult;
}

model RecommendExperiencesRequest {
  @encodedName("application/json", "required_weapons")
  requiredWeapons: {
    primary: WeaponRef;
    secondary: WeaponRef[];
  };

  limit?: int32;
}

model ExperienceRecommendation {
  id: string;
  title: string;
  category: string;

  @encodedName("application/json", "period_start")
  periodStart?: string;

  @encodedName("application/json", "period_end")
  periodEnd?: string;

  @encodedName("application/json", "star_situation")
  starSituation: string;

  weapons: string[];

  @encodedName("application/json", "match_score")
  matchScore: float32;
}

model ReviewRequest {
  @encodedName("application/json", "cover_letter_id")
  coverLetterId: string;

  content: string;
}

model ReviewScores {
  specificity: float32;

  @encodedName("application/json", "job_fit")
  jobFit: float32;

  @encodedName("application/json", "company_fit")
  companyFit: float32;

  authenticity: float32;
}

model DimensionFeedback {
  dimension: string;
  score: float32;
  good: string[];
  improve: string[];
}

model SpecificSuggestion {
  original: string;
  suggested: string;
  reason: string;
}

model ReviewResult {
  scores: ReviewScores;
  overall: float32;

  @encodedName("application/json", "per_dimension_feedback")
  perDimensionFeedback: DimensionFeedback[];

  @encodedName("application/json", "specific_suggestions")
  specificSuggestions: SpecificSuggestion[];
}

model CoachingSession {
  id: string;

  @encodedName("application/json", "session_type")
  sessionType: string;

  @encodedName("application/json", "input_tokens")
  inputTokens: int32;

  @encodedName("application/json", "output_tokens")
  outputTokens: int32;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model CoverLetter {
  id: string;

  @encodedName("application/json", "question_text")
  questionText: string;

  @encodedName("application/json", "char_limit")
  charLimit: int32;

  @encodedName("application/json", "current_content")
  currentContent: string;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;

  @encodedName("application/json", "updated_at")
  updatedAt: utcDateTime;
}

model CoverLetterVersion {
  id: string;

  @encodedName("application/json", "version_number")
  versionNumber: int32;

  content: string;

  @encodedName("application/json", "char_count")
  charCount: int32;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

model ApplicationSummary {
  id: string;

  @encodedName("application/json", "company_name")
  companyName: string;

  position: string;
  status: string;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;
}

// ─── Coaching Interface ───

@tag("Coaching")
@useAuth(BearerAuth)
interface CoachingAPI {
  /** Analyze cover letter question */
  @post
  @route("/v1/coaching/question-analysis")
  analyzeQuestion(@body body: QuestionAnalysisRequest): {
    @statusCode statusCode: 200;
    @body body: QuestionAnalysisResult;
  } | Colight.Common.ErrorResponse;

  /** Recommend experiences based on required weapons */
  @post
  @route("/v1/coaching/recommend-experiences")
  recommendExperiences(@body body: RecommendExperiencesRequest): {
    @statusCode statusCode: 200;
    @body body: { recommendations: ExperienceRecommendation[] };
  } | Colight.Common.ErrorResponse;

  /** Generate draft coaching */
  @post
  @route("/v1/coaching/draft")
  generateDraft(@body body: GenerateDraftRequest): {
    @statusCode statusCode: 200;
    @body body: { draft: string };
  } | Colight.Common.ErrorResponse;

  /** Review cover letter */
  @post
  @route("/v1/coaching/review")
  review(@body body: ReviewRequest): {
    @statusCode statusCode: 200;
    @body body: ReviewResult;
  } | Colight.Common.ErrorResponse;

  /** Get coaching sessions for a cover letter */
  @get
  @route("/v1/coaching/sessions")
  getSessions(@query @encodedName("application/json", "cover_letter_id") coverLetterId: string): {
    @statusCode statusCode: 200;
    @body body: { sessions: CoachingSession[] };
  } | Colight.Common.ErrorResponse;

  /** Get cover letter */
  @get
  @route("/v1/coaching/cover-letters/{id}")
  getCoverLetter(@path id: string): {
    @statusCode statusCode: 200;
    @body body: CoverLetter;
  } | Colight.Common.ErrorResponse;

  /** Update cover letter content */
  @patch(#{implicitOptionality: true})
  @route("/v1/coaching/cover-letters/{id}")
  updateCoverLetter(@path id: string, @body body: { content: string }): {
    @statusCode statusCode: 200;
    @body body: { @encodedName("application/json", "updated_at") updatedAt: utcDateTime };
  } | Colight.Common.ErrorResponse;

  /** Create cover letter version */
  @post
  @route("/v1/coaching/cover-letters/{id}/versions")
  createVersion(@path id: string, @body body: { content: string }): {
    @statusCode statusCode: 201;
    @body body: { version: CoverLetterVersion };
  } | Colight.Common.ErrorResponse;

  /** Get cover letter versions */
  @get
  @route("/v1/coaching/cover-letters/{id}/versions")
  getVersions(@path id: string): {
    @statusCode statusCode: 200;
    @body body: { versions: CoverLetterVersion[] };
  } | Colight.Common.ErrorResponse;

  /** Get applications list */
  @get
  @route("/v1/applications")
  getApplications(): {
    @statusCode statusCode: 200;
    @body body: { applications: ApplicationSummary[] };
  } | Colight.Common.ErrorResponse;
}

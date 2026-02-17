import { apiClient } from "@/lib/api-client";

// === Types ===

export interface QuestionAnalysisRequest {
  application_id?: string;
  company_name?: string;
  question_text: string;
  char_limit: number;
}

export interface QuestionAnalysisResult {
  surface_question: string;
  real_intents: Array<{
    intent: string;
    why: string;
  }>;
  required_weapons: {
    primary: { weapon_id: string; weapon_name: string; reason: string };
    secondary: Array<{
      weapon_id: string;
      weapon_name: string;
      reason: string;
    }>;
  };
  writing_structure: {
    total_chars: number;
    sections: Array<{
      name: string;
      char_ratio: number;
      char_count: number;
      guide: string;
    }>;
  };
  key_keywords: string[];
  avoid_list: string[];
  good_structure_example: string;
}

export interface GenerateDraftRequest {
  application_id?: string;
  company_name?: string;
  experience_ids: string[];
  question_text: string;
  char_limit: number;
  analysis_result?: QuestionAnalysisResult;
}

export interface GenerateDraftResponse {
  draft: string;
}

export interface CoachingSession {
  id: string;
  session_type: string;
  input_tokens: number;
  output_tokens: number;
  created_at: string;
}

export interface CoverLetter {
  id: string;
  question_text: string;
  char_limit: number;
  current_content: string;
  created_at: string;
  updated_at: string;
}

export interface CoverLetterVersion {
  id: string;
  version_number: number;
  content: string;
  char_count: number;
  change_summary?: string;
  created_at: string;
}

// === Review Types ===

export interface ReviewScores {
  specificity: number;
  job_fit: number;
  company_fit: number;
  authenticity: number;
}

export interface DimensionFeedback {
  dimension: string;
  score: number;
  good: string[];
  improve: string[];
}

export interface SpecificSuggestion {
  original: string;
  suggested: string;
  reason: string;
}

export interface ReviewResult {
  scores: ReviewScores;
  overall: number;
  per_dimension_feedback: DimensionFeedback[];
  specific_suggestions: SpecificSuggestion[];
}

export interface ReviewRequest {
  cover_letter_id: string;
  content: string;
}

// === Char Coaching Types ===

export interface CharCoachingSuggestion {
  type: "trim" | "expand";
  section: string;
  original: string;
  suggested: string;
  reason: string;
  char_diff: number;
}

export interface CharCoachingResult {
  status: "over" | "under" | "good";
  current_count: number;
  char_limit: number;
  diff: number;
  suggestions: CharCoachingSuggestion[];
  summary: string;
}

export interface CharCoachingRequest {
  cover_letter_id: string;
  content: string;
}

// === Advice Types ===

export interface AdviceItem {
  category: string; // metric, structure, detail, keyword
  content: string;
  priority: number; // 1=high, 2=medium, 3=low
}

// === API Functions ===

export interface ExperienceRecommendation {
  id: string;
  title: string;
  category: string;
  period_start?: string;
  period_end?: string;
  star_situation: string;
  weapons: string[];
  match_score: number;
  match_reasons: string[];
  is_used: boolean;
  keyword_matches: string[];
}

export async function recommendExperiences(
  requiredWeapons: QuestionAnalysisResult["required_weapons"],
  keyKeywords: string[] = [],
  applicationId = "",
  limit = 10,
): Promise<{ recommendations: ExperienceRecommendation[] }> {
  const { data } = await apiClient.post<{
    recommendations: ExperienceRecommendation[];
  }>("/v1/coaching/recommend-experiences", {
    required_weapons: requiredWeapons,
    key_keywords: keyKeywords,
    application_id: applicationId,
    limit,
  });
  return data;
}

export async function analyzeQuestion(
  req: QuestionAnalysisRequest
): Promise<QuestionAnalysisResult> {
  const { data } = await apiClient.post<QuestionAnalysisResult>(
    "/v1/coaching/question-analysis",
    req
  );
  return data;
}

export async function generateDraft(
  req: GenerateDraftRequest
): Promise<GenerateDraftResponse> {
  const { data } = await apiClient.post<GenerateDraftResponse>(
    "/v1/coaching/draft",
    req
  );
  return data;
}

export async function getCoachingSessions(
  coverLetterId: string
): Promise<{ sessions: CoachingSession[] }> {
  const { data } = await apiClient.get<{ sessions: CoachingSession[] }>(
    "/v1/coaching/sessions",
    { params: { cover_letter_id: coverLetterId } }
  );
  return data;
}

export async function getCoverLetter(id: string): Promise<CoverLetter> {
  const { data } = await apiClient.get<CoverLetter>(
    `/v1/coaching/cover-letters/${id}`
  );
  return data;
}

export async function updateCoverLetter(
  id: string,
  content: string
): Promise<{ updated_at: string }> {
  const { data } = await apiClient.patch<{ updated_at: string }>(
    `/v1/coaching/cover-letters/${id}`,
    { content }
  );
  return data;
}

export async function createVersion(
  id: string,
  content: string,
  changeSummary?: string
): Promise<{ version: CoverLetterVersion }> {
  const body: Record<string, string> = { content };
  if (changeSummary) {
    body.change_summary = changeSummary;
  }
  const { data } = await apiClient.post<{ version: CoverLetterVersion }>(
    `/v1/coaching/cover-letters/${id}/versions`,
    body
  );
  return data;
}

export async function getVersions(
  id: string
): Promise<{ versions: CoverLetterVersion[] }> {
  const { data } = await apiClient.get<{ versions: CoverLetterVersion[] }>(
    `/v1/coaching/cover-letters/${id}/versions`
  );
  return data;
}

export async function requestReview(
  req: ReviewRequest
): Promise<ReviewResult> {
  const { data } = await apiClient.post<ReviewResult>(
    "/v1/coaching/review",
    req
  );
  return data;
}

export async function requestCharCoaching(
  req: CharCoachingRequest
): Promise<CharCoachingResult> {
  const { data } = await apiClient.post<CharCoachingResult>(
    "/v1/coaching/char-count",
    req
  );
  return data;
}

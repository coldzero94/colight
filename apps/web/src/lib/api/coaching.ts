import { apiClient } from "@/lib/api-client";

// === Types ===

export interface QuestionAnalysisRequest {
  application_id: string;
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
  application_id: string;
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
  created_at: string;
}

export interface ApplicationSummary {
  id: string;
  company_name: string;
  position: string;
  status: string;
  created_at: string;
}

// === API Functions ===

export async function getApplications(): Promise<{
  applications: ApplicationSummary[];
}> {
  const { data } = await apiClient.get<{
    applications: ApplicationSummary[];
  }>("/v1/applications");
  return data;
}

export interface ExperienceRecommendation {
  id: string;
  title: string;
  category: string;
  period_start?: string;
  period_end?: string;
  star_situation: string;
  weapons: string[];
  match_score: number;
}

export async function recommendExperiences(
  requiredWeapons: QuestionAnalysisResult["required_weapons"],
  limit = 10
): Promise<{ recommendations: ExperienceRecommendation[] }> {
  const { data } = await apiClient.post<{
    recommendations: ExperienceRecommendation[];
  }>("/v1/coaching/recommend-experiences", {
    required_weapons: requiredWeapons,
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
  content: string
): Promise<{ version: CoverLetterVersion }> {
  const { data } = await apiClient.post<{ version: CoverLetterVersion }>(
    `/v1/coaching/cover-letters/${id}/versions`,
    { content }
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

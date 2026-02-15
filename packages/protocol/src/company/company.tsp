import "@typespec/http";
import "../common/errors.tsp";

using TypeSpec.Http;

namespace Colight.Company;

// ─── Models ───

model CoreValue {
  keyword: string;
  description: string;
}

model TalentTrait {
  trait: string;
  description: string;
  evidence?: string;
}

model Trend {
  title: string;
  summary: string;
  relevance?: string;
}

model CompanyAnalysis {
  @encodedName("application/json", "company_name")
  companyName: string;

  @encodedName("application/json", "core_values")
  coreValues: CoreValue[];

  @encodedName("application/json", "talent_traits")
  talentTraits: TalentTrait[];

  @encodedName("application/json", "recent_trends")
  recentTrends: Trend[];

  @encodedName("application/json", "strategy_keywords")
  strategyKeywords: string[];

  @encodedName("application/json", "avoid_expressions")
  avoidExpressions: string[];

  source: "talent_profiles" | "cache" | "ai_generated";
}

model AnalyzeCompanyRequest {
  @encodedName("application/json", "company_name")
  companyName: string;
}

model BasicInfo {
  @encodedName("application/json", "corp_name")
  corpName: string;

  @encodedName("application/json", "corp_code")
  corpCode?: string;

  @encodedName("application/json", "stock_code")
  stockCode?: string;

  ceo?: string;
  industry?: string;
  address?: string;
}

model NewsItem {
  title: string;
  link: string;
  description?: string;

  @encodedName("application/json", "pub_date")
  pubDate?: string;

  source?: string;
}

model CompanyData {
  @encodedName("application/json", "basic_info")
  basicInfo: BasicInfo;

  news: NewsItem[];
}

model MatchResult {
  @encodedName("application/json", "experience_id")
  experienceId: string;

  @encodedName("application/json", "overall_fit")
  overallFit: float32;

  @encodedName("application/json", "job_relevance")
  jobRelevance: float32;

  @encodedName("application/json", "talent_fit")
  talentFit: float32;

  uniqueness: float32;
  reasoning: string;

  @encodedName("application/json", "suggested_angle")
  suggestedAngle: string;
}

model MatchRequest {
  @encodedName("application/json", "company_name")
  companyName: string;
}

model MatchResponse {
  @encodedName("application/json", "company_name")
  companyName: string;

  matches: MatchResult[];
  total: int32;
}

// ─── Company Interface ───

@tag("Company")
@useAuth(BearerAuth)
interface CompanyAPI {
  /** Analyze company (core values, talent traits, trends) */
  @post
  @route("/v1/analyze-company")
  analyzeCompany(@body body: AnalyzeCompanyRequest): {
    @statusCode statusCode: 200;
    @body body: CompanyAnalysis;
  } | Colight.Common.ErrorResponse;

  /** Get company data (DART + News) */
  @get
  @route("/v1/company-data")
  getCompanyData(@query name: string): {
    @statusCode statusCode: 200;
    @body body: CompanyData;
  } | Colight.Common.ErrorResponse;

  /** Match experiences against company */
  @post
  @route("/v1/match")
  matchExperiences(@body body: MatchRequest): {
    @statusCode statusCode: 200;
    @body body: MatchResponse;
  } | Colight.Common.ErrorResponse;
}

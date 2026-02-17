import { apiClient } from "@/lib/api-client";

// Company analysis types
export interface CoreValue {
  keyword: string;
  description: string;
}

export interface TalentTrait {
  trait: string;
  description: string;
  evidence?: string;
}

export interface Trend {
  title: string;
  summary: string;
  relevance?: string;
}

export interface NewsArticle {
  title: string;
  link: string;
  description?: string;
  pub_date?: string;
}

export interface CompanyAnalysis {
  company_name: string;
  core_values: CoreValue[];
  talent_traits: TalentTrait[];
  recent_trends: Trend[];
  strategy_keywords: string[];
  avoid_expressions: string[];
  source: "talent_profiles" | "cache" | "ai_generated";
  cached_at?: string;
  view_count?: number;
  source_news?: NewsArticle[];
}

// Analyze company
export async function analyzeCompany(
  companyName: string,
  jobPostingUrl?: string
): Promise<CompanyAnalysis> {
  const { data } = await apiClient.post<CompanyAnalysis>("/v1/analyze-company", {
    company_name: companyName,
    job_posting_url: jobPostingUrl,
  });
  return data;
}

// Get company data (DART + News)
export interface CompanyData {
  basic_info: {
    corp_name: string;
    corp_code?: string;
    stock_code?: string;
    ceo?: string;
    industry?: string;
    address?: string;
  };
  news: Array<{
    title: string;
    link: string;
    description?: string;
    pub_date?: string;
    source?: string;
  }>;
}

export async function getCompanyData(companyName: string): Promise<CompanyData> {
  const { data } = await apiClient.get<CompanyData>("/v1/company-data", {
    params: { name: companyName },
  });
  return data;
}

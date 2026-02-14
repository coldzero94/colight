import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import {
  useCompanyData,
  useAnalyzeCompany,
  useCompanyAnalysis,
} from "../use-company-analysis";

vi.mock("@/lib/api/analysis", () => ({
  analyzeCompany: vi.fn(),
  getCompanyData: vi.fn(),
}));

import { analyzeCompany, getCompanyData } from "@/lib/api/analysis";

const mockAnalyzeCompany = vi.mocked(analyzeCompany);
const mockGetCompanyData = vi.mocked(getCompanyData);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  const Wrapper = ({ children }: { children: React.ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
  Wrapper.displayName = "TestQueryWrapper";
  return Wrapper;
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useCompanyData", () => {
  it("fetches company data when name is provided", async () => {
    mockGetCompanyData.mockResolvedValue({
      basic_info: { corp_name: "삼성전자" },
      news: [],
    });

    const { result } = renderHook(() => useCompanyData("삼성전자"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockGetCompanyData).toHaveBeenCalledWith("삼성전자");
    expect(result.current.data?.basic_info.corp_name).toBe("삼성전자");
  });

  it("does not fetch when name is null", () => {
    const { result } = renderHook(() => useCompanyData(null), {
      wrapper: createWrapper(),
    });

    expect(result.current.isFetching).toBe(false);
    expect(mockGetCompanyData).not.toHaveBeenCalled();
  });
});

describe("useAnalyzeCompany", () => {
  it("calls analyzeCompany mutation", async () => {
    const mockResult = {
      company_name: "LG전자",
      core_values: [{ keyword: "혁신", description: "기술 혁신" }],
      talent_traits: [],
      recent_trends: [],
      strategy_keywords: ["혁신"],
      avoid_expressions: [],
      source: "ai_generated" as const,
    };
    mockAnalyzeCompany.mockResolvedValue(mockResult);

    const { result } = renderHook(() => useAnalyzeCompany(), {
      wrapper: createWrapper(),
    });

    result.current.mutate("LG전자");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAnalyzeCompany).toHaveBeenCalledWith("LG전자");
    expect(result.current.data?.company_name).toBe("LG전자");
  });

  it("handles mutation error", async () => {
    mockAnalyzeCompany.mockRejectedValue(new Error("API error"));

    const { result } = renderHook(() => useAnalyzeCompany(), {
      wrapper: createWrapper(),
    });

    result.current.mutate("실패회사");

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeDefined();
  });
});

describe("useCompanyAnalysis", () => {
  it("fetches analysis when name is provided", async () => {
    const mockResult = {
      company_name: "카카오",
      core_values: [],
      talent_traits: [],
      recent_trends: [],
      strategy_keywords: [],
      avoid_expressions: [],
      source: "cache" as const,
    };
    mockAnalyzeCompany.mockResolvedValue(mockResult);

    const { result } = renderHook(() => useCompanyAnalysis("카카오"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.company_name).toBe("카카오");
    expect(result.current.data?.source).toBe("cache");
  });

  it("does not fetch when name is null", () => {
    const { result } = renderHook(() => useCompanyAnalysis(null), {
      wrapper: createWrapper(),
    });

    expect(result.current.isFetching).toBe(false);
  });
});

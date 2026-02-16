import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { CoachingFlow } from "../coaching-flow";

vi.mock("@/lib/api/coaching", () => ({
  analyzeQuestion: vi.fn(),
  recommendExperiences: vi.fn(),
}));

vi.mock("@/hooks/use-draft-streaming", () => ({
  useDraftStreaming: () => ({
    content: "",
    isStreaming: false,
    error: null,
    coverLetterId: null,
    sessionId: null,
    streamDraft: vi.fn(),
    abort: vi.fn(),
  }),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn() }),
}));

const mockUseApplications = vi.fn();
vi.mock("@/hooks/use-applications", () => ({
  useApplications: (...args: unknown[]) => mockUseApplications(...args),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("CoachingFlow", () => {
  it("shows loading state while fetching applications", () => {
    mockUseApplications.mockReturnValue({ data: undefined, isLoading: true });
    render(<CoachingFlow />, { wrapper: createWrapper() });
    expect(screen.getByText("불러오는 중...")).toBeInTheDocument();
  });

  it("shows empty state when no applications exist", async () => {
    mockUseApplications.mockReturnValue({ data: [], isLoading: false });
    render(<CoachingFlow />, { wrapper: createWrapper() });
    await waitFor(() =>
      expect(screen.getByText("먼저 기업 분석을 진행해주세요")).toBeInTheDocument()
    );
  });

  it("shows question form when applications exist", async () => {
    mockUseApplications.mockReturnValue({
      data: [
        {
          id: "a0000000-0000-0000-0000-000000000001",
          company_name: "삼성전자",
          position: "백엔드 개발자",
          status: "preparing",
          deadline: null,
          applied_at: null,
          notes: "",
          cover_letter_count: 0,
          created_at: "2026-01-01T00:00:00Z",
          updated_at: "2026-01-01T00:00:00Z",
        },
      ],
      isLoading: false,
    });
    render(<CoachingFlow />, { wrapper: createWrapper() });
    await waitFor(() =>
      expect(screen.getByText("분석 시작")).toBeInTheDocument()
    );
  });

  it("renders company select with application data", async () => {
    mockUseApplications.mockReturnValue({
      data: [
        {
          id: "a0000000-0000-0000-0000-000000000001",
          company_name: "삼성전자",
          position: "백엔드 개발자",
          status: "preparing",
          deadline: null,
          applied_at: null,
          notes: "",
          cover_letter_count: 0,
          created_at: "2026-01-01T00:00:00Z",
          updated_at: "2026-01-01T00:00:00Z",
        },
        {
          id: "a0000000-0000-0000-0000-000000000002",
          company_name: "네이버",
          position: "프론트엔드 개발자",
          status: "preparing",
          deadline: null,
          applied_at: null,
          notes: "",
          cover_letter_count: 0,
          created_at: "2026-01-02T00:00:00Z",
          updated_at: "2026-01-02T00:00:00Z",
        },
      ],
      isLoading: false,
    });
    render(<CoachingFlow />, { wrapper: createWrapper() });
    await waitFor(() =>
      expect(
        screen.getByText("삼성전자 - 백엔드 개발자")
      ).toBeInTheDocument()
    );
    expect(
      screen.getByText("네이버 - 프론트엔드 개발자")
    ).toBeInTheDocument();
  });
});

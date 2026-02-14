import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { EditorPageClient } from "../editor-page-client";
import * as coachingApi from "@/lib/api/coaching";

vi.mock("@/lib/api/coaching", () => ({
  getCoverLetter: vi.fn(),
  getVersions: vi.fn(),
  updateCoverLetter: vi.fn(),
  createVersion: vi.fn(),
}));

vi.mock("next/link", () => ({
  default: ({
    children,
    href,
  }: {
    children: React.ReactNode;
    href: string;
  }) => <a href={href}>{children}</a>,
}));

const mockGetCoverLetter = vi.mocked(coachingApi.getCoverLetter);
const mockGetVersions = vi.mocked(coachingApi.getVersions);

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
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

describe("EditorPageClient", () => {
  it("shows loading state initially", () => {
    mockGetCoverLetter.mockReturnValue(new Promise(() => {}));
    mockGetVersions.mockResolvedValue({ versions: [] });

    render(<EditorPageClient coverLetterId="cl-123" />, {
      wrapper: createWrapper(),
    });
    expect(
      screen.getByText("에디터를 불러오는 중...")
    ).toBeInTheDocument();
  });

  it("shows error state when cover letter not found", async () => {
    mockGetCoverLetter.mockRejectedValue(new Error("Not found"));
    mockGetVersions.mockResolvedValue({ versions: [] });

    render(<EditorPageClient coverLetterId="cl-123" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() =>
      expect(
        screen.getByText("자소서를 불러올 수 없습니다")
      ).toBeInTheDocument()
    );
  });

  it("renders editor when data loads successfully", async () => {
    mockGetCoverLetter.mockResolvedValue({
      id: "cl-123",
      question_text: "테스트 문항",
      char_limit: 800,
      current_content: "초안 내용",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    });
    mockGetVersions.mockResolvedValue({ versions: [] });

    render(<EditorPageClient coverLetterId="cl-123" />, {
      wrapper: createWrapper(),
    });

    await waitFor(() =>
      expect(screen.getByText("테스트 문항")).toBeInTheDocument()
    );
  });
});

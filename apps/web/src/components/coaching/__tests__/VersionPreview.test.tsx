import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VersionPreview } from "../version-preview";

const mockVersion = {
  id: "v-1",
  version_number: 2,
  content: "이전 버전 내용입니다",
  char_count: 11,
  created_at: "2026-02-14T10:00:00Z",
};

const currentContent = "현재 내용입니다";

describe("VersionPreview", () => {
  it("renders version metadata", () => {
    render(
      <VersionPreview
        version={mockVersion}
        currentContent={currentContent}
        onRestore={vi.fn()}
        onClose={vi.fn()}
      />
    );
    expect(screen.getByText("v2")).toBeInTheDocument();
    expect(screen.getByText("11자")).toBeInTheDocument();
  });

  it("shows version content by default", () => {
    render(
      <VersionPreview
        version={mockVersion}
        currentContent={currentContent}
        onRestore={vi.fn()}
        onClose={vi.fn()}
      />
    );
    expect(screen.getByText("이전 버전 내용입니다")).toBeInTheDocument();
  });

  it("shows change_summary when present", () => {
    const versionWithSummary = {
      ...mockVersion,
      change_summary: "v1에서 복원",
    };
    render(
      <VersionPreview
        version={versionWithSummary}
        currentContent={currentContent}
        onRestore={vi.fn()}
        onClose={vi.fn()}
      />
    );
    expect(screen.getByText(/v1에서 복원/)).toBeInTheDocument();
  });

  it("toggles diff view on button click", async () => {
    const user = userEvent.setup();
    render(
      <VersionPreview
        version={mockVersion}
        currentContent={currentContent}
        onRestore={vi.fn()}
        onClose={vi.fn()}
      />
    );

    await user.click(screen.getByText("비교 보기"));

    // Diff stats should be visible
    expect(screen.getByText(/추가/)).toBeInTheDocument();
    expect(screen.getByText(/삭제/)).toBeInTheDocument();
  });

  it("calls onClose when back button clicked", async () => {
    const user = userEvent.setup();
    const handleClose = vi.fn();
    render(
      <VersionPreview
        version={mockVersion}
        currentContent={currentContent}
        onRestore={vi.fn()}
        onClose={handleClose}
      />
    );

    await user.click(screen.getByText("돌아가기"));
    expect(handleClose).toHaveBeenCalled();
  });

  it("calls onRestore when restore button clicked", async () => {
    const user = userEvent.setup();
    const handleRestore = vi.fn();
    render(
      <VersionPreview
        version={mockVersion}
        currentContent={currentContent}
        onRestore={handleRestore}
        onClose={vi.fn()}
      />
    );

    await user.click(screen.getByText("이 버전으로 복원"));
    expect(handleRestore).toHaveBeenCalled();
  });
});

import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VersionHistory } from "../version-history";

describe("VersionHistory", () => {
  const mockVersions = [
    {
      id: "v-2",
      version_number: 2,
      content: "v2 content",
      char_count: 650,
      created_at: "2026-02-14T11:00:00Z",
    },
    {
      id: "v-1",
      version_number: 1,
      content: "v1 content",
      char_count: 500,
      created_at: "2026-02-14T10:00:00Z",
    },
  ];

  it("renders nothing when no versions", () => {
    const { container } = render(
      <VersionHistory versions={[]} coverLetterId="cl-1" />
    );
    expect(container.innerHTML).toBe("");
  });

  it("shows version count in collapsed state", () => {
    render(<VersionHistory versions={mockVersions} coverLetterId="cl-1" />);
    expect(screen.getByText("버전 이력 (2개)")).toBeInTheDocument();
  });

  it("expands to show version details on click", async () => {
    const user = userEvent.setup();
    render(<VersionHistory versions={mockVersions} coverLetterId="cl-1" />);

    await user.click(screen.getByText("버전 이력 (2개)"));

    expect(screen.getByText("v1")).toBeInTheDocument();
    expect(screen.getByText("500자")).toBeInTheDocument();
    expect(screen.getByText("v2")).toBeInTheDocument();
    expect(screen.getByText("650자")).toBeInTheDocument();
  });

  it("shows (현재) label on latest version", async () => {
    const user = userEvent.setup();
    render(<VersionHistory versions={mockVersions} coverLetterId="cl-1" />);

    await user.click(screen.getByText("버전 이력 (2개)"));

    expect(screen.getByText("(현재)")).toBeInTheDocument();
  });

  it("calls onSelectVersion when a version is clicked", async () => {
    const user = userEvent.setup();
    const handleSelect = vi.fn();
    render(
      <VersionHistory
        versions={mockVersions}
        coverLetterId="cl-1"
        onSelectVersion={handleSelect}
      />
    );

    await user.click(screen.getByText("버전 이력 (2개)"));
    await user.click(screen.getByText("v1"));

    expect(handleSelect).toHaveBeenCalledWith(mockVersions[1]);
  });

  it("highlights selected version", async () => {
    const user = userEvent.setup();
    render(
      <VersionHistory
        versions={mockVersions}
        coverLetterId="cl-1"
        selectedVersionId="v-1"
      />
    );

    await user.click(screen.getByText("버전 이력 (2개)"));

    const v1Button = screen.getByText("v1").closest("button");
    expect(v1Button).toHaveClass("bg-blue-50");
    expect(v1Button).toHaveClass("border-blue-200");
  });

  it("shows change_summary when present", async () => {
    const user = userEvent.setup();
    const versionsWithSummary = [
      {
        ...mockVersions[0],
        change_summary: "v1에서 복원",
      },
      mockVersions[1],
    ];
    render(
      <VersionHistory
        versions={versionsWithSummary}
        coverLetterId="cl-1"
      />
    );

    await user.click(screen.getByText("버전 이력 (2개)"));

    expect(screen.getByText("v1에서 복원")).toBeInTheDocument();
  });
});

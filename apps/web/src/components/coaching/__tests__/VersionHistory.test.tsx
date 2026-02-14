import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { VersionHistory } from "../version-history";

describe("VersionHistory", () => {
  const mockVersions = [
    {
      id: "v-1",
      version_number: 1,
      content: "v1 content",
      char_count: 500,
      created_at: "2026-02-14T10:00:00Z",
    },
    {
      id: "v-2",
      version_number: 2,
      content: "v2 content",
      char_count: 650,
      created_at: "2026-02-14T11:00:00Z",
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
});

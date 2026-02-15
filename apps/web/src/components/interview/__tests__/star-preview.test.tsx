import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { STARPreview } from "../star-preview";
import type { ExtractSTARResult } from "@/lib/api/interview";

const mockData: ExtractSTARResult = {
  title: "팀 프로젝트 리더",
  category: "project",
  content: "대학교 팀 프로젝트를 이끌었습니다.",
  result: "A+ 학점을 받았습니다.",
  star_situation: "4학년 캡스톤 프로젝트",
  star_task: "3개월 내 프로토타입 완성",
  star_action: "매주 스프린트 회의를 진행",
  star_result: "기한 내 완성하여 A+ 학점",
  keywords: ["리더십", "팀워크"],
};

describe("STARPreview", () => {
  it("renders all STAR fields", () => {
    render(
      <STARPreview
        data={mockData}
        isSaving={false}
        onSave={vi.fn()}
        onReExtract={vi.fn()}
      />
    );

    expect(screen.getByDisplayValue("팀 프로젝트 리더")).toBeInTheDocument();
    expect(screen.getByDisplayValue("4학년 캡스톤 프로젝트")).toBeInTheDocument();
    expect(
      screen.getByDisplayValue("3개월 내 프로토타입 완성")
    ).toBeInTheDocument();
    expect(
      screen.getByDisplayValue("매주 스프린트 회의를 진행")
    ).toBeInTheDocument();
    expect(
      screen.getByDisplayValue("기한 내 완성하여 A+ 학점")
    ).toBeInTheDocument();
    expect(screen.getByText("리더십")).toBeInTheDocument();
    expect(screen.getByText("팀워크")).toBeInTheDocument();
  });

  it("calls onSave with edited data", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();

    render(
      <STARPreview
        data={mockData}
        isSaving={false}
        onSave={onSave}
        onReExtract={vi.fn()}
      />
    );

    // Edit the title
    const titleInput = screen.getByDisplayValue("팀 프로젝트 리더");
    await user.clear(titleInput);
    await user.type(titleInput, "수정된 제목");

    // Click save
    await user.click(screen.getByRole("button", { name: /저장하기/ }));

    expect(onSave).toHaveBeenCalledWith(
      expect.objectContaining({ title: "수정된 제목" })
    );
  });

  it("calls onReExtract when re-extract button is clicked", async () => {
    const user = userEvent.setup();
    const onReExtract = vi.fn();

    render(
      <STARPreview
        data={mockData}
        isSaving={false}
        onSave={vi.fn()}
        onReExtract={onReExtract}
      />
    );

    await user.click(screen.getByRole("button", { name: /다시 추출/ }));
    expect(onReExtract).toHaveBeenCalled();
  });

  it("shows saving spinner when isSaving", () => {
    render(
      <STARPreview
        data={mockData}
        isSaving={true}
        onSave={vi.fn()}
        onReExtract={vi.fn()}
      />
    );

    expect(screen.getByText("저장 중...")).toBeInTheDocument();
  });
});

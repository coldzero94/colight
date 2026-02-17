import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AdvicePanel } from "../advice-panel";
import type { AdviceItem } from "@/lib/api/coaching";

const mockAdvice: AdviceItem[] = [
  { category: "metric", content: "성과를 수치로 표현하면 설득력이 높아집니다.", priority: 1 },
  { category: "structure", content: "STAR 구조의 결과 부분을 보강하세요.", priority: 2 },
  { category: "keyword", content: "'데이터 기반' 키워드를 추가하세요.", priority: 3 },
];

describe("AdvicePanel", () => {
  it("renders all advice items", () => {
    render(<AdvicePanel advice={mockAdvice} />);

    expect(screen.getByText(/수치로 표현/)).toBeInTheDocument();
    expect(screen.getByText(/결과 부분을 보강/)).toBeInTheDocument();
    expect(screen.getByText(/데이터 기반/)).toBeInTheDocument();
  });

  it("renders category labels in Korean", () => {
    render(<AdvicePanel advice={mockAdvice} />);

    expect(screen.getByText("수치/데이터 보강")).toBeInTheDocument();
    expect(screen.getByText("구조 개선")).toBeInTheDocument();
    expect(screen.getByText("키워드 활용")).toBeInTheDocument();
  });

  it("sorts items by priority (high first)", () => {
    const reversed: AdviceItem[] = [
      { category: "keyword", content: "키워드 조언", priority: 3 },
      { category: "metric", content: "수치 조언", priority: 1 },
    ];
    render(<AdvicePanel advice={reversed} />);

    const items = screen.getAllByRole("listitem");
    expect(items[0]).toHaveTextContent("수치 조언");
    expect(items[1]).toHaveTextContent("키워드 조언");
  });

  it("allows dismissing items via checkbox", async () => {
    const user = userEvent.setup();
    render(<AdvicePanel advice={mockAdvice} />);

    const checkboxes = screen.getAllByRole("checkbox");
    expect(checkboxes).toHaveLength(3);

    // Click first checkbox
    await user.click(checkboxes[0]);

    // First item should have line-through style
    expect(checkboxes[0]).toBeChecked();
  });

  it("renders nothing when advice is empty", () => {
    const { container } = render(<AdvicePanel advice={[]} />);
    expect(container.firstChild).toBeNull();
  });
});

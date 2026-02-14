import { EditorLayout } from "@/components/coaching/editor-layout";

// Mock data for now - will be replaced with actual API calls
const mockCoverLetter = {
  id: "test-id",
  question_text: "팀 프로젝트에서 어려움을 극복한 경험을 구체적으로 기술하세요",
  char_limit: 800,
  current_content: "[상황]\n초안 내용이 여기에 표시됩니다...",
  company_name: "테스트회사",
  position: "백엔드 개발자",
};

const mockAnalysis = {
  required_weapons: {
    primary: { weapon_id: "W01", weapon_name: "문제해결", reason: "핵심 역량" },
    secondary: [{ weapon_id: "W02", weapon_name: "협업", reason: "팀 경험" }],
  },
  writing_structure: {
    total_chars: 800,
    sections: [
      { name: "상황", char_ratio: 0.2, char_count: 160, guide: "상황 설정" },
      { name: "과제", char_ratio: 0.15, char_count: 120, guide: "과제 정의" },
      { name: "행동", char_ratio: 0.4, char_count: 320, guide: "실행 과정" },
      { name: "결과", char_ratio: 0.25, char_count: 200, guide: "성과" },
    ],
  },
  key_keywords: ["데이터 기반", "개선율", "주도적"],
};

const mockExperiences = [
  {
    id: "exp-1",
    title: "프로젝트 경험",
    category: "프로젝트",
    star_situation: "팀 프로젝트 상황",
    weapons: [{ name: "문제해결" }, { name: "협업" }],
    matchScore: 92,
  },
];

export default async function CoachingEditPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  // TODO: Fetch actual data from backend API
  // const coverLetter = await fetch(`/v1/coaching/cover-letters/${id}`);

  const handleSave = async (content: string) => {
    "use server";
    // TODO: Call backend PATCH /v1/coaching/cover-letters/:id
    console.log("Saving content:", content);
  };

  return (
    <EditorLayout
      coverLetter={mockCoverLetter}
      analysis={mockAnalysis}
      experiences={mockExperiences}
      onSave={handleSave}
    />
  );
}

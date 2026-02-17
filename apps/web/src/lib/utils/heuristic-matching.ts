import type { Experience } from "@/lib/api/experiences";
import type { CompanyAnalysis } from "@/lib/api/analysis";

export interface HeuristicMatch {
  experience_id: string;
  overall_fit: number;
  weapon_score: number;
  keyword_score: number;
  category_score: number;
  reasoning: string;
}

/**
 * Fast heuristic matching between experiences and company analysis.
 * No AI calls - instant results based on keyword/weapon overlap.
 */
export function matchExperiencesHeuristic(
  experiences: Experience[],
  analysis: CompanyAnalysis
): HeuristicMatch[] {
  if (!experiences || experiences.length === 0) return [];
  if (!analysis) return [];

  const matches = experiences.map((exp) => {
    const scores = calculateMatchScore(exp, analysis);
    return {
      experience_id: exp.id,
      overall_fit: Math.round(scores.overall),
      weapon_score: Math.round(scores.weapon),
      keyword_score: Math.round(scores.keyword),
      category_score: Math.round(scores.category),
      reasoning: generateReasoning(exp, scores, analysis),
    };
  });

  // Sort by overall_fit descending
  return matches.sort((a, b) => b.overall_fit - a.overall_fit);
}

function calculateMatchScore(exp: Experience, analysis: CompanyAnalysis) {
  // 1. Weapon matching (50% weight)
  const weaponScore = calculateWeaponScore(exp, analysis);

  // 2. Keyword matching (30% weight)
  const keywordScore = calculateKeywordScore(exp, analysis);

  // 3. Category matching (20% weight)
  const categoryScore = calculateCategoryScore(exp);

  const overall = weaponScore * 0.5 + keywordScore * 0.3 + categoryScore * 0.2;

  return {
    weapon: weaponScore,
    keyword: keywordScore,
    category: categoryScore,
    overall,
  };
}

function calculateWeaponScore(exp: Experience, analysis: CompanyAnalysis): number {
  if (!exp.weapons || exp.weapons.length === 0) return 0;
  if (!analysis.talent_traits || analysis.talent_traits.length === 0) return 50;

  // Map weapon codes to keywords for matching
  const weaponKeywords: Record<string, string[]> = {
    W01: ["위기", "극복", "실패", "역경", "도전"],
    W02: ["리더", "주도", "이끌", "팀장", "책임"],
    W03: ["협업", "팀워크", "갈등", "조율", "소통"],
    W04: ["도전", "목표", "달성", "성취", "혁신"],
    W05: ["문제", "해결", "분석", "개선", "창의"],
    W06: ["소통", "설득", "발표", "협상", "경청"],
    W07: ["성장", "학습", "배움", "전문", "발전"],
  };

  const talentText = analysis.talent_traits
    .map((t) => (t.trait || "") + " " + (t.description || ""))
    .join(" ")
    .toLowerCase();

  let matchCount = 0;
  for (const weapon of exp.weapons) {
    const keywords = weaponKeywords[weapon.weapon_code] || [];
    const hasMatch = keywords.some((kw) => talentText.includes(kw));
    if (hasMatch) matchCount++;
  }

  // 50-100 scale based on weapon coverage
  const matchRatio = matchCount / exp.weapons.length;
  return 50 + matchRatio * 50;
}

function calculateKeywordScore(exp: Experience, analysis: CompanyAnalysis): number {
  if (!analysis.strategy_keywords || analysis.strategy_keywords.length === 0)
    return 50;

  const expContent = [
    exp.title,
    exp.content,
    exp.star_situation,
    exp.star_task,
    exp.star_action,
    exp.star_result,
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();

  const matches = analysis.strategy_keywords.filter((keyword) =>
    expContent.includes(keyword.toLowerCase())
  );

  // 0-100 scale based on keyword coverage
  const coverage = matches.length / analysis.strategy_keywords.length;
  return coverage * 100;
}

function calculateCategoryScore(exp: Experience): number {
  // Work experience scores higher than other categories
  const categoryScores: Record<string, number> = {
    work: 100,
    project: 85,
    activity: 70,
    competition: 75,
    education: 60,
    volunteer: 65,
    other: 50,
  };

  return categoryScores[exp.category] ?? 50;
}

function generateReasoning(
  exp: Experience,
  scores: ReturnType<typeof calculateMatchScore>,
  analysis: CompanyAnalysis
): string {
  const reasons: string[] = [];

  if (scores.weapon >= 75) {
    reasons.push("역량 무기가 인재상과 잘 부합합니다");
  } else if (scores.weapon >= 50) {
    reasons.push("역량 무기가 부분적으로 일치합니다");
  }

  if (scores.keyword >= 60) {
    const matchCount = Math.floor(
      (scores.keyword / 100) * (analysis.strategy_keywords?.length || 0)
    );
    reasons.push(`전략 키워드 ${matchCount}개 포함`);
  }

  if (exp.category === "work") {
    reasons.push("실무 경험");
  }

  return reasons.join(". ") || "기본적인 적합도를 보입니다";
}

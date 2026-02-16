"use client";

interface AnalysisResult {
  required_weapons: {
    primary: { weapon_id: string; weapon_name: string; reason: string };
    secondary: Array<{ weapon_id: string; weapon_name: string; reason: string }>;
  };
  writing_structure: {
    total_chars: number;
    sections: Array<{
      name: string;
      char_ratio: number;
      char_count: number;
      guide: string;
    }>;
  };
  key_keywords: string[];
}

interface Experience {
  id: string;
  title: string;
  category: string;
  star_situation: string;
  weapons: Array<{ name: string }>;
  matchScore: number;
}

interface AnalysisSidebarProps {
  analysis: AnalysisResult;
  experiences: Experience[];
}

export function AnalysisSidebar({ analysis, experiences }: AnalysisSidebarProps) {
  return (
    <div className="space-y-6 rounded-lg border border-border bg-card p-6">
      <h3 className="text-lg font-semibold text-foreground">📊 분석 요약</h3>

      {/* 필요 무기 */}
      <div>
        <h4 className="mb-2 text-sm font-medium text-foreground/80">필요 무기</h4>
        <div className="flex flex-wrap gap-2">
          <span className="inline-flex items-center gap-1 rounded-full bg-yellow-500/10 px-3 py-1 text-sm font-semibold text-yellow-400">
            🏆 {analysis.required_weapons.primary.weapon_name}
          </span>
          {analysis.required_weapons.secondary.map((weapon, i) => (
            <span
              key={i}
              className="rounded-full bg-white/[0.06] px-3 py-1 text-sm font-medium text-foreground/80"
            >
              {weapon.weapon_name}
            </span>
          ))}
        </div>
      </div>

      {/* 추천 구조 */}
      <div>
        <h4 className="mb-2 text-sm font-medium text-foreground/80">추천 구조</h4>
        <div className="space-y-1 text-xs">
          {analysis.writing_structure.sections.map((section, i) => (
            <div key={i} className="flex justify-between text-muted-foreground">
              <span>{section.name}</span>
              <span>
                {Math.round(section.char_ratio * 100)}% ({section.char_count}자)
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* 핵심 키워드 */}
      <div>
        <h4 className="mb-2 text-sm font-medium text-foreground/80">핵심 키워드</h4>
        <div className="flex flex-wrap gap-1">
          {analysis.key_keywords.map((keyword, i) => (
            <span
              key={i}
              className="rounded bg-primary/10 px-2 py-1 text-xs font-medium text-primary"
            >
              {keyword}
            </span>
          ))}
        </div>
      </div>

      {/* 사용 경험 */}
      <div>
        <h4 className="mb-2 text-sm font-medium text-foreground/80">사용 경험</h4>
        <div className="space-y-2">
          {experiences.map((exp) => (
            <div
              key={exp.id}
              className="rounded-lg border border-border bg-background p-3"
            >
              <div className="flex items-start justify-between">
                <h5 className="text-sm font-semibold text-foreground">
                  {exp.title}
                </h5>
                <span className="text-xs font-medium text-blue-400">
                  {exp.matchScore}%
                </span>
              </div>
              <div className="mt-1 flex flex-wrap gap-1">
                {exp.weapons.map((weapon, i) => (
                  <span
                    key={i}
                    className="rounded bg-white/[0.06] px-1.5 py-0.5 text-xs text-muted-foreground"
                  >
                    {weapon.name}
                  </span>
                ))}
              </div>
              <p className="mt-2 line-clamp-2 text-xs text-muted-foreground">
                {exp.star_situation}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

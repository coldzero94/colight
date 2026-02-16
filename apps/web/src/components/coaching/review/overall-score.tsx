"use client";

interface OverallScoreProps {
  score: number;
  previousScore?: number;
}

function getGrade(score: number): { label: string; color: string } {
  if (score >= 90) return { label: "S", color: "text-purple-400" };
  if (score >= 80) return { label: "A", color: "text-blue-400" };
  if (score >= 70) return { label: "B", color: "text-green-400" };
  if (score >= 60) return { label: "C", color: "text-yellow-400" };
  return { label: "D", color: "text-red-400" };
}

export function OverallScore({ score, previousScore }: OverallScoreProps) {
  const grade = getGrade(score);
  const diff = previousScore != null ? score - previousScore : null;

  return (
    <div className="flex items-center gap-4 rounded-lg border border-border bg-card p-6">
      <div className="flex flex-col items-center">
        <span className={`text-4xl font-bold ${grade.color}`}>{score}</span>
        <span className="text-xs text-muted-foreground/60">/ 100</span>
      </div>
      <div className="flex flex-col">
        <span className={`text-2xl font-bold ${grade.color}`}>
          {grade.label}
        </span>
        {diff != null && (
          <span
            className={`text-sm font-medium ${
              diff > 0
                ? "text-green-400"
                : diff < 0
                  ? "text-red-400"
                  : "text-muted-foreground/60"
            }`}
          >
            {diff > 0 ? `+${diff}` : diff === 0 ? "→" : diff}
          </span>
        )}
      </div>
    </div>
  );
}

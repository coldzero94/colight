import type { MatchResult } from "@/lib/api/matching";
import { WeaponBadges } from "@/components/experiences/weapon-badges";
import type { Experience } from "@/lib/api/experiences";

interface MatchingResultsProps {
  matches: MatchResult[];
  experiences: Experience[];
}

export function MatchingResults({ matches, experiences }: MatchingResultsProps) {
  if (matches.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        매칭 결과가 없습니다.
      </div>
    );
  }

  // Create experience map for quick lookup
  const expMap = new Map(experiences.map((e) => [e.id, e]));

  // Sort by overall_fit descending
  const sortedMatches = [...matches].sort((a, b) => b.overall_fit - a.overall_fit);

  return (
    <div className="space-y-4">
      {sortedMatches.map((match) => {
        const exp = expMap.get(match.experience_id);
        if (!exp) return null;

        return (
          <div
            key={match.experience_id}
            className="rounded-lg border border-gray-200 bg-white p-6 hover:border-gray-300 transition-colors"
          >
            {/* Header */}
            <div className="mb-4 flex items-start justify-between">
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-gray-900">
                  {exp.title}
                </h3>
                {exp.weapons && exp.weapons.length > 0 && (
                  <div className="mt-2">
                    <WeaponBadges weapons={exp.weapons} maxVisible={3} />
                  </div>
                )}
              </div>
              <div className="ml-4">
                <FitScoreBadge score={match.overall_fit} />
              </div>
            </div>

            {/* Scores breakdown */}
            <div className="mb-4 grid grid-cols-3 gap-4">
              <ScoreBar
                label="직무 관련도"
                score={match.job_relevance}
                color="blue"
                weight={40}
              />
              <ScoreBar
                label="인재상 부합"
                score={match.talent_fit}
                color="green"
                weight={35}
              />
              <ScoreBar
                label="차별화"
                score={match.uniqueness}
                color="purple"
                weight={25}
              />
            </div>

            {/* Reasoning */}
            <div className="mb-3">
              <h4 className="text-sm font-medium text-gray-700 mb-1">
                매칭 근거
              </h4>
              <p className="text-sm text-gray-600">{match.reasoning}</p>
            </div>

            {/* Suggested angle */}
            <div className="rounded-lg bg-blue-50 p-3">
              <h4 className="text-sm font-medium text-blue-900 mb-1">
                💡 활용 제안
              </h4>
              <p className="text-sm text-blue-700">{match.suggested_angle}</p>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function FitScoreBadge({ score }: { score: number }) {
  let bgColor = "bg-gray-100";
  let textColor = "text-gray-700";

  if (score >= 80) {
    bgColor = "bg-green-100";
    textColor = "text-green-700";
  } else if (score >= 60) {
    bgColor = "bg-blue-100";
    textColor = "text-blue-700";
  } else if (score >= 40) {
    bgColor = "bg-yellow-100";
    textColor = "text-yellow-700";
  } else {
    bgColor = "bg-red-100";
    textColor = "text-red-700";
  }

  return (
    <div
      className={`inline-flex items-center rounded-full ${bgColor} ${textColor} px-4 py-2 text-sm font-semibold`}
    >
      {score}점
    </div>
  );
}

interface ScoreBarProps {
  label: string;
  score: number;
  color: "blue" | "green" | "purple";
  weight: number;
}

function ScoreBar({ label, score, color, weight }: ScoreBarProps) {
  const colors = {
    blue: "bg-blue-500",
    green: "bg-green-500",
    purple: "bg-purple-500",
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs text-gray-600">
          {label} ({weight}%)
        </span>
        <span className="text-xs font-medium text-gray-900">{score}</span>
      </div>
      <div className="h-2 w-full rounded-full bg-gray-100">
        <div
          className={`h-2 rounded-full ${colors[color]}`}
          style={{ width: `${score}%` }}
        />
      </div>
    </div>
  );
}

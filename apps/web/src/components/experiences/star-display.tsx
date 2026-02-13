interface StarDisplayProps {
  situation?: string;
  task?: string;
  action?: string;
  result?: string;
}

const sections = [
  { key: "situation", label: "Situation", icon: "🔍", description: "상황/배경" },
  { key: "task", label: "Task", icon: "🎯", description: "과제/목표" },
  { key: "action", label: "Action", icon: "⚡", description: "구체적 행동" },
  { key: "result", label: "Result", icon: "📊", description: "결과/성과" },
] as const;

export function StarDisplay({ situation, task, action, result }: StarDisplayProps) {
  const values = { situation, task, action, result };

  return (
    <div className="space-y-4" data-testid="star-display">
      {sections.map((section) => {
        const content = values[section.key];
        return (
          <div
            key={section.key}
            className="rounded-lg border border-gray-100 bg-white p-4"
          >
            <h4 className="mb-2 flex items-center gap-1.5 text-sm font-semibold text-gray-900">
              <span>{section.icon}</span>
              <span>{section.label}</span>
              <span className="text-xs font-normal text-gray-400">
                {section.description}
              </span>
            </h4>
            {content ? (
              <p className="text-sm leading-relaxed text-gray-700 whitespace-pre-wrap">
                {content}
              </p>
            ) : (
              <p className="text-sm text-gray-400" data-testid="star-empty">
                아직 작성되지 않았습니다
              </p>
            )}
          </div>
        );
      })}
    </div>
  );
}

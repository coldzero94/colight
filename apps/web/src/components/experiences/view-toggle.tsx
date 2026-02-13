"use client";

interface ViewToggleProps {
  mode: "grid" | "list";
  onChange: (mode: "grid" | "list") => void;
}

export function ViewToggle({ mode, onChange }: ViewToggleProps) {
  return (
    <div className="flex rounded-lg border border-gray-200" role="group" aria-label="뷰 전환">
      <button
        type="button"
        onClick={() => onChange("grid")}
        className={`px-3 py-1.5 text-sm font-medium rounded-l-lg transition-colors ${
          mode === "grid"
            ? "bg-gray-900 text-white"
            : "bg-white text-gray-600 hover:bg-gray-50"
        }`}
        aria-pressed={mode === "grid"}
        aria-label="그리드 뷰"
      >
        ▦
      </button>
      <button
        type="button"
        onClick={() => onChange("list")}
        className={`px-3 py-1.5 text-sm font-medium rounded-r-lg transition-colors ${
          mode === "list"
            ? "bg-gray-900 text-white"
            : "bg-white text-gray-600 hover:bg-gray-50"
        }`}
        aria-pressed={mode === "list"}
        aria-label="리스트 뷰"
      >
        ☰
      </button>
    </div>
  );
}

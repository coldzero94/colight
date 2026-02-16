"use client";

interface SortSelectProps {
  value: string;
  onChange: (value: string) => void;
}

const sortOptions = [
  { value: "latest", label: "최신순" },
  { value: "oldest", label: "오래된순" },
  { value: "title", label: "제목순" },
];

export function SortSelect({ value, onChange }: SortSelectProps) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="rounded-lg border border-border bg-input px-3 py-1.5 text-sm text-foreground/80 focus:outline-none focus:ring-2 focus:ring-ring"
      aria-label="정렬"
    >
      {sortOptions.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  );
}

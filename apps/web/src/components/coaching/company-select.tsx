"use client";

interface Company {
  id: string;
  company_name: string;
  position: string;
  created_at: string;
}

interface CompanySelectProps {
  companies: Company[];
  value: string;
  onChange: (value: string) => void;
  error?: string;
}

export function CompanySelect({
  companies,
  value,
  onChange,
  error,
}: CompanySelectProps) {
  return (
    <div className="space-y-1">
      <label htmlFor="company" className="text-sm font-medium text-foreground/80">
        기업 선택
      </label>
      <select
        id="company"
        role="combobox"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full rounded-lg border border-border bg-input px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
      >
        <option value="">기업을 선택하세요</option>
        {companies.map((company) => (
          <option key={company.id} value={company.id} role="option">
            {company.company_name} - {company.position}
          </option>
        ))}
      </select>
      {error && (
        <p role="alert" className="text-xs text-red-500">
          {error}
        </p>
      )}
    </div>
  );
}

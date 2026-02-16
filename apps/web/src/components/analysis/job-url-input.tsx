"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  jobUrlSchema,
  detectDomain,
  type JobUrlFormValues,
} from "@/lib/validations/job-url";

interface JobUrlInputProps {
  onSubmit: (url: string) => void;
  isLoading: boolean;
  error?: string;
}

export function JobUrlInput({ onSubmit, isLoading, error }: JobUrlInputProps) {
  const [url, setUrl] = useState("");
  const domainInfo = url ? detectDomain(url) : null;

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<JobUrlFormValues>({
    resolver: zodResolver(jobUrlSchema),
  });

  const handleFormSubmit = (data: JobUrlFormValues) => {
    onSubmit(data.url);
  };

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm font-medium text-foreground/80">
          채용공고 URL
        </label>
        <div className="relative">
          <input
            {...register("url")}
            type="text"
            placeholder="채용공고 URL을 입력하세요 (예: https://www.jobkorea.co.kr/...)"
            className="w-full rounded-lg border border-border bg-background px-4 py-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
            onChange={(e) => setUrl(e.target.value)}
            disabled={isLoading}
          />
          {domainInfo && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2">
              <DomainBadge domain={domainInfo.domain} label={domainInfo.label} />
            </div>
          )}
        </div>
        {errors.url && (
          <p className="text-xs text-red-400">{errors.url.message}</p>
        )}
        {error && <p className="text-xs text-red-500">{error}</p>}
        {domainInfo && !domainInfo.supported && (
          <p className="text-xs text-amber-400">
            {domainInfo.label}는 현재 지원 예정입니다. AI 자동 분석으로 처리됩니다.
          </p>
        )}
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50 transition-colors"
      >
        {isLoading ? "분석 중..." : "분석 시작"}
      </button>
    </form>
  );
}

interface DomainBadgeProps {
  domain: string;
  label: string;
}

function DomainBadge({ domain, label }: DomainBadgeProps) {
  const colors = {
    jobkorea: "bg-blue-500/10 text-blue-400",
    catch: "bg-primary/10 text-primary",
    wanted: "bg-purple-500/10 text-purple-400",
    saramin: "bg-red-500/10 text-red-400",
    unknown: "bg-white/[0.06] text-foreground/80",
  };

  const colorClass = colors[domain as keyof typeof colors] || colors.unknown;

  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${colorClass}`}
    >
      {label}
    </span>
  );
}

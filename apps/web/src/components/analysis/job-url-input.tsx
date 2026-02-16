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
            placeholder="채용공고 URL을 입력하세요"
            className="brand-input w-full rounded-xl px-4 py-3 text-sm"
            onChange={(e) => setUrl(e.target.value)}
            disabled={isLoading}
          />
          {domainInfo && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2">
              <span className="inline-flex items-center rounded-full border border-primary/35 bg-primary/15 px-2 py-1 text-xs font-medium text-primary">
                {domainInfo.label}
              </span>
            </div>
          )}
        </div>
        {errors.url && (
          <p className="text-xs text-red-400">{errors.url.message}</p>
        )}
        {error && <p className="text-xs text-red-500">{error}</p>}
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full rounded-xl bg-primary px-4 py-3 text-sm font-medium text-primary-foreground shadow-[0_12px_28px_rgba(72,132,255,0.26)] transition-all hover:-translate-y-0.5 hover:bg-primary/90 disabled:opacity-50"
      >
        {isLoading ? "분석 중..." : "분석 시작"}
      </button>
    </form>
  );
}

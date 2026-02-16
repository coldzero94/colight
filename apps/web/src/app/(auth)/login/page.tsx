"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";
import { loginSchema, type LoginFormValues } from "@/lib/auth";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000";

export default function LoginPage() {
  const router = useRouter();
  const { setTokens, fetchUser } = useAuthStore();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (values: LoginFormValues) => {
    setIsSubmitting(true);
    try {
      const { data } = await apiClient.post("/v1/auth/login", values);
      setTokens(data.tokens.access_token, data.tokens.refresh_token);
      await fetchUser();
      router.replace("/experiences");
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } } };
      const message =
        error.response?.data?.error?.message || "로그인에 실패했습니다.";
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleNaverLogin = () => {
    window.location.href = `${API_URL}/v1/auth/naver/login`;
  };

  return (
    <div className="space-y-6">
      <button
        type="button"
        onClick={handleNaverLogin}
        className="w-full flex items-center justify-center gap-2 bg-[#03C75A] hover:bg-[#02b351] text-white font-medium py-3 px-4 rounded-lg transition-all duration-200 shadow-md shadow-[#03C75A]/20"
      >
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
          <path
            d="M13.5 10.5L6.2 3H3v14h3.5V9.5L13.8 17H17V3h-3.5v7.5z"
            fill="currentColor"
          />
        </svg>
        Naver로 시작하기
      </button>

      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-border" />
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="bg-card px-4 text-muted-foreground">또는</span>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label
            htmlFor="email"
            className="block text-sm font-medium text-foreground/80 mb-1.5"
          >
            이메일
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            {...register("email")}
            className="w-full px-3.5 py-2.5 border border-border bg-white/[0.06] rounded-lg text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 transition-all duration-200"
            placeholder="example@email.com"
          />
          {errors.email && (
            <p className="mt-1.5 text-sm text-destructive">{errors.email.message}</p>
          )}
        </div>

        <div>
          <label
            htmlFor="password"
            className="block text-sm font-medium text-foreground/80 mb-1.5"
          >
            비밀번호
          </label>
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            {...register("password")}
            className="w-full px-3.5 py-2.5 border border-border bg-white/[0.06] rounded-lg text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/50 transition-all duration-200"
            placeholder="8자 이상"
          />
          {errors.password && (
            <p className="mt-1.5 text-sm text-destructive">
              {errors.password.message}
            </p>
          )}
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full bg-primary hover:bg-primary/90 disabled:opacity-50 text-primary-foreground font-medium py-3 px-4 rounded-lg transition-all duration-200 shadow-md shadow-primary/20"
        >
          {isSubmitting ? "로그인 중..." : "로그인"}
        </button>
      </form>

      <p className="text-center text-sm text-muted-foreground">
        계정이 없으신가요?{" "}
        <Link href="/signup" className="text-primary hover:underline">
          회원가입
        </Link>
      </p>
    </div>
  );
}

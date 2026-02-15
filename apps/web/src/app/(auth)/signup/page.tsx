"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";
import { signupSchema, type SignupFormValues } from "@/lib/auth";

export default function SignupPage() {
  const router = useRouter();
  const { setTokens, fetchUser } = useAuthStore();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SignupFormValues>({
    resolver: zodResolver(signupSchema),
  });

  const onSubmit = async (values: SignupFormValues) => {
    setIsSubmitting(true);
    try {
      const { data } = await apiClient.post("/v1/auth/signup", {
        email: values.email,
        password: values.password,
        nickname: values.nickname || undefined,
      });
      setTokens(data.tokens.access_token, data.tokens.refresh_token);
      await fetchUser();
      router.replace("/experiences");
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } } };
      const message =
        error.response?.data?.error?.message || "회원가입에 실패했습니다.";
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900 text-center">
        회원가입
      </h2>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label
            htmlFor="email"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            이메일
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            {...register("email")}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="example@email.com"
          />
          {errors.email && (
            <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
          )}
        </div>

        <div>
          <label
            htmlFor="password"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            비밀번호
          </label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            {...register("password")}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="영문 + 숫자 조합, 8자 이상"
          />
          {errors.password && (
            <p className="mt-1 text-sm text-red-600">
              {errors.password.message}
            </p>
          )}
        </div>

        <div>
          <label
            htmlFor="passwordConfirm"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            비밀번호 확인
          </label>
          <input
            id="passwordConfirm"
            type="password"
            autoComplete="new-password"
            {...register("passwordConfirm")}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="비밀번호를 다시 입력해주세요"
          />
          {errors.passwordConfirm && (
            <p className="mt-1 text-sm text-red-600">
              {errors.passwordConfirm.message}
            </p>
          )}
        </div>

        <div>
          <label
            htmlFor="nickname"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            닉네임 <span className="text-gray-400">(선택)</span>
          </label>
          <input
            id="nickname"
            type="text"
            {...register("nickname")}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="최대 50자"
          />
          {errors.nickname && (
            <p className="mt-1 text-sm text-red-600">
              {errors.nickname.message}
            </p>
          )}
        </div>

        <div className="flex items-start gap-2">
          <input
            id="agreeToTerms"
            type="checkbox"
            {...register("agreeToTerms")}
            className="mt-1 h-4 w-4 rounded border-gray-300"
          />
          <label htmlFor="agreeToTerms" className="text-sm text-gray-600">
            <Link
              href="/terms"
              target="_blank"
              className="text-blue-600 hover:underline"
            >
              이용약관
            </Link>
            {" 및 "}
            <Link
              href="/privacy"
              target="_blank"
              className="text-blue-600 hover:underline"
            >
              개인정보처리방침
            </Link>
            에 동의합니다.
          </label>
        </div>
        {errors.agreeToTerms && (
          <p className="text-sm text-red-600">
            {errors.agreeToTerms.message}
          </p>
        )}

        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full bg-gray-900 hover:bg-gray-800 disabled:bg-gray-400 text-white font-medium py-3 px-4 rounded-lg transition-colors"
        >
          {isSubmitting ? "가입 중..." : "회원가입"}
        </button>
      </form>

      <p className="text-center text-sm text-gray-600">
        이미 계정이 있으신가요?{" "}
        <Link href="/login" className="text-blue-600 hover:underline">
          로그인
        </Link>
      </p>
    </div>
  );
}

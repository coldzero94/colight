"use client";

import Link from "next/link";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:9000";

export default function SignupPage() {
  const handleNaverLogin = () => {
    window.location.href = `${API_URL}/v1/auth/naver/login`;
  };

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-foreground text-center">
        회원가입
      </h2>

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

      <p className="text-center text-sm text-muted-foreground">
        이미 계정이 있으신가요?{" "}
        <Link href="/login" className="text-primary hover:underline">
          로그인
        </Link>
      </p>
    </div>
  );
}

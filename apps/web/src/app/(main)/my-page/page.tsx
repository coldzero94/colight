"use client";

import { useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";

export default function MyPage() {
  const { user, refreshUser } = useAuthStore();
  const [loading, setLoading] = useState(false);

  // Profile update
  const [nickname, setNickname] = useState(user?.nickname || "");

  // Password change
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickname.trim()) {
      toast.error("닉네임을 입력해주세요.");
      return;
    }
    setLoading(true);
    try {
      await apiClient.put("/v1/profile", { nickname });
      await refreshUser();
      toast.success("프로필이 수정되었습니다.");
    } catch {
      toast.error("프로필 수정에 실패했습니다.");
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentPassword || !newPassword) {
      toast.error("모든 필드를 입력해주세요.");
      return;
    }
    if (newPassword.length < 8) {
      toast.error("새 비밀번호는 최소 8자 이상이어야 합니다.");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("새 비밀번호가 일치하지 않습니다.");
      return;
    }
    setLoading(true);
    try {
      await apiClient.post("/v1/profile/password", {
        current_password: currentPassword,
        new_password: newPassword,
      });
      toast.success("비밀번호가 변경되었습니다.");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err: any) {
      const msg = err.response?.data?.error?.message || "비밀번호 변경에 실패했습니다.";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold font-display text-foreground">
          마이페이지
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          프로필 정보와 비밀번호를 관리하세요
        </p>
      </div>

      {/* Profile Update */}
      <div className="bg-card border border-border rounded-xl p-6 space-y-4">
        <h2 className="text-lg font-semibold text-foreground">프로필 수정</h2>
        <form onSubmit={handleUpdateProfile} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              이메일
            </label>
            <input
              type="text"
              value={user?.email || ""}
              disabled
              className="w-full px-3 py-2 border border-border rounded-lg bg-muted text-muted-foreground cursor-not-allowed"
            />
            <p className="text-xs text-muted-foreground mt-1">
              이메일은 변경할 수 없습니다
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">
              닉네임
            </label>
            <input
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              maxLength={50}
              className="w-full px-3 py-2 border border-border rounded-lg bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
          >
            {loading ? "저장 중..." : "프로필 저장"}
          </button>
        </form>
      </div>

      {/* Password Change (Email accounts only) */}
      {user?.auth_provider === "email" && (
        <div className="bg-card border border-border rounded-xl p-6 space-y-4">
          <h2 className="text-lg font-semibold text-foreground">비밀번호 변경</h2>
          <form onSubmit={handleChangePassword} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                현재 비밀번호
              </label>
              <input
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                className="w-full px-3 py-2 border border-border rounded-lg bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                새 비밀번호
              </label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full px-3 py-2 border border-border rounded-lg bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
              <p className="text-xs text-muted-foreground mt-1">
                최소 8자 이상
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-2">
                새 비밀번호 확인
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full px-3 py-2 border border-border rounded-lg bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
            >
              {loading ? "변경 중..." : "비밀번호 변경"}
            </button>
          </form>
        </div>
      )}

      {user?.auth_provider !== "email" && (
        <div className="bg-card border border-border rounded-xl p-6">
          <p className="text-sm text-muted-foreground">
            {user?.auth_provider === "naver" ? "네이버" : "소셜"} 로그인 계정은 비밀번호를 변경할 수 없습니다.
          </p>
        </div>
      )}
    </div>
  );
}

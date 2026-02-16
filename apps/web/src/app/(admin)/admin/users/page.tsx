"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";
import type { UserRole } from "@/lib/auth-utils";
import { hasRole } from "@/lib/auth-utils";

interface AdminUser {
  id: string;
  email?: string;
  nickname?: string;
  auth_provider: "email" | "naver";
  role: UserRole;
  plan?: string;
  suspended?: boolean;
  last_login_at?: string;
  created_at: string;
}

const ROLE_BADGE_STYLES: Record<UserRole, string> = {
  user: "bg-white/[0.06] text-foreground/80",
  manager: "bg-blue-500/10 text-blue-400",
  admin: "bg-purple-500/10 text-purple-400",
  super_admin: "bg-red-500/10 text-red-400",
};

const ROLE_LABELS: Record<UserRole, string> = {
  user: "User",
  manager: "Manager",
  admin: "Admin",
  super_admin: "Super Admin",
};

const PLAN_LABELS: Record<string, string> = {
  free: "Free",
  starter: "Starter",
  pro: "Pro",
  season: "Season",
};

const ALL_ROLES: UserRole[] = ["user", "manager", "admin", "super_admin"];
const ALL_PLANS = ["free", "starter", "pro", "season"];

const PAGE_SIZE = 20;

export default function AdminUsersPage() {
  const { user: currentUser } = useAuthStore();
  const isSuperAdmin = currentUser?.role === "super_admin";
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [roleFilter, setRoleFilter] = useState("");
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);

  // Detail drawer
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [suspendReason, setSuspendReason] = useState("");
  const [deleteReason, setDeleteReason] = useState("");
  const [planValue, setPlanValue] = useState("");

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        limit: String(PAGE_SIZE),
        offset: String(page * PAGE_SIZE),
      });
      if (roleFilter) params.set("role", roleFilter);
      if (search) params.set("search", search);

      const { data } = await apiClient.get(`/v1/admin/users?${params}`);
      setUsers(data.data);
      setTotal(data.total);
    } catch {
      toast.error("사용자 목록을 불러올 수 없습니다.");
    } finally {
      setLoading(false);
    }
  }, [page, roleFilter, search]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const handleRoleChange = async (userId: string, newRole: UserRole) => {
    if (!confirm(`역할을 '${ROLE_LABELS[newRole]}'로 변경하시겠습니까?`))
      return;

    try {
      await apiClient.put(`/v1/admin/users/${userId}/role`, { role: newRole });
      toast.success("역할이 변경되었습니다.");
      fetchUsers();
    } catch {
      toast.error("역할 변경에 실패했습니다.");
    }
  };

  const handleSuspend = async () => {
    if (!selectedUser) return;
    setActionLoading(true);
    try {
      await apiClient.post(`/v1/admin/users/${selectedUser.id}/suspend`, {
        reason: suspendReason,
      });
      toast.success("계정이 정지되었습니다.");
      setSuspendReason("");
      fetchUsers();
      setSelectedUser((prev) =>
        prev ? { ...prev, suspended: true } : null
      );
    } catch {
      toast.error("계정 정지에 실패했습니다.");
    } finally {
      setActionLoading(false);
    }
  };

  const handleUnsuspend = async () => {
    if (!selectedUser) return;
    setActionLoading(true);
    try {
      await apiClient.delete(`/v1/admin/users/${selectedUser.id}/suspend`);
      toast.success("정지가 해제되었습니다.");
      fetchUsers();
      setSelectedUser((prev) =>
        prev ? { ...prev, suspended: false } : null
      );
    } catch {
      toast.error("정지 해제에 실패했습니다.");
    } finally {
      setActionLoading(false);
    }
  };

  const handleForceLogout = async () => {
    if (!selectedUser) return;
    if (!confirm("이 사용자를 강제 로그아웃 시키시겠습니까?")) return;
    setActionLoading(true);
    try {
      await apiClient.post(
        `/v1/admin/users/${selectedUser.id}/force-logout`
      );
      toast.success("강제 로그아웃 처리되었습니다.");
    } catch {
      toast.error("강제 로그아웃에 실패했습니다.");
    } finally {
      setActionLoading(false);
    }
  };

  const handlePlanChange = async () => {
    if (!selectedUser || !planValue) return;
    if (
      !confirm(
        `플랜을 '${PLAN_LABELS[planValue]}'로 변경하시겠습니까?`
      )
    )
      return;
    setActionLoading(true);
    try {
      await apiClient.put(`/v1/admin/users/${selectedUser.id}/plan`, {
        plan: planValue,
      });
      toast.success("플랜이 변경되었습니다.");
      fetchUsers();
      setSelectedUser((prev) =>
        prev ? { ...prev, plan: planValue } : null
      );
    } catch {
      toast.error("플랜 변경에 실패했습니다.");
    } finally {
      setActionLoading(false);
    }
  };

  const handleExport = async () => {
    if (!selectedUser) return;
    setActionLoading(true);
    try {
      const { data } = await apiClient.post(
        `/v1/admin/users/${selectedUser.id}/export`
      );
      const blob = new Blob([JSON.stringify(data, null, 2)], {
        type: "application/json",
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `user-export-${selectedUser.id}.json`;
      a.click();
      URL.revokeObjectURL(url);
      toast.success("사용자 데이터가 다운로드되었습니다.");
    } catch {
      toast.error("데이터 내보내기에 실패했습니다.");
    } finally {
      setActionLoading(false);
    }
  };

  const handleCreateDeletion = async () => {
    if (!selectedUser) return;
    if (!confirm("이 사용자에 대해 삭제 요청을 생성하시겠습니까? (30일 후 삭제)"))
      return;
    setActionLoading(true);
    try {
      await apiClient.post(
        `/v1/admin/users/${selectedUser.id}/delete-request`,
        { reason: deleteReason }
      );
      toast.success("삭제 요청이 생성되었습니다.");
      setDeleteReason("");
    } catch {
      toast.error("삭제 요청 생성에 실패했습니다.");
    } finally {
      setActionLoading(false);
    }
  };

  const openDetail = (user: AdminUser) => {
    setSelectedUser(user);
    setPlanValue(user.plan ?? "free");
    setSuspendReason("");
    setDeleteReason("");
  };

  // Roles the current user can assign (strictly below their own level)
  const assignableRoles = currentUser
    ? ALL_ROLES.filter(
        (r) =>
          r !== currentUser.role && !hasRole(r as UserRole, currentUser.role)
      )
    : [];

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold font-display text-foreground">
          사용자 관리
        </h1>
        <p className="text-sm text-muted-foreground">총 {total}명</p>
      </div>

      <div className="flex gap-3">
        <input
          type="text"
          placeholder="이메일 또는 닉네임 검색"
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setPage(0);
          }}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        />
        <select
          value={roleFilter}
          onChange={(e) => {
            setRoleFilter(e.target.value);
            setPage(0);
          }}
          className="px-3 py-2 border border-border rounded-lg text-sm bg-card text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        >
          <option value="">전체 역할</option>
          {ALL_ROLES.map((role) => (
            <option key={role} value={role}>
              {ROLE_LABELS[role]}
            </option>
          ))}
        </select>
      </div>

      <div className="flex gap-6">
        {/* Users Table */}
        <div className="flex-1 bg-card border border-border rounded-xl overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-white/[0.02] border-b border-border">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  이메일
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  닉네임
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  역할
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  플랜
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  상태
                </th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">
                  가입일
                </th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td
                    colSpan={7}
                    className="text-center py-8 text-muted-foreground/60"
                  >
                    로딩 중...
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td
                    colSpan={7}
                    className="text-center py-8 text-muted-foreground/60"
                  >
                    사용자가 없습니다.
                  </td>
                </tr>
              ) : (
                users.map((user) => {
                  const isSelf = user.id === currentUser?.id;
                  const canChange =
                    !isSelf &&
                    currentUser &&
                    hasRole(currentUser.role, user.role) &&
                    currentUser.role !== user.role;

                  return (
                    <tr
                      key={user.id}
                      className={`border-b border-border hover:bg-white/[0.04] cursor-pointer ${
                        selectedUser?.id === user.id
                          ? "bg-primary/5"
                          : ""
                      }`}
                      onClick={() => openDetail(user)}
                    >
                      <td className="px-4 py-3">{user.email ?? "-"}</td>
                      <td className="px-4 py-3">{user.nickname ?? "-"}</td>
                      <td className="px-4 py-3">
                        {canChange ? (
                          <select
                            value={user.role}
                            onClick={(e) => e.stopPropagation()}
                            onChange={(e) =>
                              handleRoleChange(
                                user.id,
                                e.target.value as UserRole
                              )
                            }
                            className="text-xs border border-border rounded px-2 py-1 bg-card text-foreground"
                          >
                            <option value={user.role}>
                              {ROLE_LABELS[user.role]}
                            </option>
                            {assignableRoles
                              .filter((r) => r !== user.role)
                              .map((role) => (
                                <option key={role} value={role}>
                                  {ROLE_LABELS[role]}
                                </option>
                              ))}
                          </select>
                        ) : (
                          <span
                            className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${ROLE_BADGE_STYLES[user.role]}`}
                          >
                            {ROLE_LABELS[user.role]}
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <span className="text-xs text-muted-foreground">
                          {PLAN_LABELS[user.plan ?? "free"] ?? "Free"}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        {user.suspended ? (
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-medium bg-red-500/10 text-red-400">
                            정지됨
                          </span>
                        ) : (
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-medium bg-green-500/10 text-green-400">
                            활성
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground">
                        {new Date(user.created_at).toLocaleDateString(
                          "ko-KR"
                        )}
                      </td>
                      <td className="px-4 py-3">
                        {isSelf && (
                          <span className="text-xs text-muted-foreground/60">
                            본인
                          </span>
                        )}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Detail Drawer */}
        {selectedUser && (
          <div className="w-80 shrink-0 bg-card border border-border rounded-xl p-5 space-y-5 self-start sticky top-4">
            <div className="flex items-center justify-between">
              <h2 className="text-base font-semibold text-foreground">
                사용자 상세
              </h2>
              <button
                onClick={() => setSelectedUser(null)}
                className="text-muted-foreground hover:text-foreground text-sm"
              >
                닫기
              </button>
            </div>

            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">이메일</span>
                <span className="text-foreground">
                  {selectedUser.email ?? "-"}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">닉네임</span>
                <span className="text-foreground">
                  {selectedUser.nickname ?? "-"}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">역할</span>
                <span
                  className={`px-2 py-0.5 rounded text-xs font-medium ${ROLE_BADGE_STYLES[selectedUser.role]}`}
                >
                  {ROLE_LABELS[selectedUser.role]}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">플랜</span>
                <span className="text-foreground">
                  {PLAN_LABELS[selectedUser.plan ?? "free"]}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">상태</span>
                {selectedUser.suspended ? (
                  <span className="px-2 py-0.5 rounded text-xs font-medium bg-red-500/10 text-red-400">
                    정지됨
                  </span>
                ) : (
                  <span className="px-2 py-0.5 rounded text-xs font-medium bg-green-500/10 text-green-400">
                    활성
                  </span>
                )}
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">ID</span>
                <span className="text-foreground/60 font-mono text-xs">
                  {selectedUser.id.slice(0, 12)}...
                </span>
              </div>
            </div>

            <div className="border-t border-border pt-4 space-y-4">
              {/* Suspend / Unsuspend */}
              {selectedUser.id !== currentUser?.id && (
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    계정 정지
                  </p>
                  {selectedUser.suspended ? (
                    <button
                      onClick={handleUnsuspend}
                      disabled={actionLoading}
                      className="w-full px-3 py-1.5 text-sm bg-green-500/10 text-green-400 rounded-lg hover:bg-green-500/20 disabled:opacity-50"
                    >
                      정지 해제
                    </button>
                  ) : (
                    <div className="space-y-2">
                      <input
                        type="text"
                        placeholder="정지 사유 (선택)"
                        value={suspendReason}
                        onChange={(e) => setSuspendReason(e.target.value)}
                        className="w-full px-3 py-1.5 text-sm border border-border rounded-lg bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
                      />
                      <button
                        onClick={handleSuspend}
                        disabled={actionLoading}
                        className="w-full px-3 py-1.5 text-sm bg-red-500/10 text-red-400 rounded-lg hover:bg-red-500/20 disabled:opacity-50"
                      >
                        계정 정지
                      </button>
                    </div>
                  )}
                </div>
              )}

              {/* Force Logout */}
              {selectedUser.id !== currentUser?.id && (
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    강제 로그아웃
                  </p>
                  <button
                    onClick={handleForceLogout}
                    disabled={actionLoading}
                    className="w-full px-3 py-1.5 text-sm bg-orange-500/10 text-orange-400 rounded-lg hover:bg-orange-500/20 disabled:opacity-50"
                  >
                    강제 로그아웃
                  </button>
                </div>
              )}

              {/* Plan Change */}
              <div className="space-y-2">
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  플랜 변경
                </p>
                <div className="flex gap-2">
                  <select
                    value={planValue}
                    onChange={(e) => setPlanValue(e.target.value)}
                    className="flex-1 px-3 py-1.5 text-sm border border-border rounded-lg bg-card text-foreground"
                  >
                    {ALL_PLANS.map((p) => (
                      <option key={p} value={p}>
                        {PLAN_LABELS[p]}
                      </option>
                    ))}
                  </select>
                  <button
                    onClick={handlePlanChange}
                    disabled={
                      actionLoading ||
                      planValue === (selectedUser.plan ?? "free")
                    }
                    className="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
                  >
                    변경
                  </button>
                </div>
              </div>

              {/* Export User Data */}
              <div className="space-y-2">
                <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  데이터 내보내기
                </p>
                <button
                  onClick={handleExport}
                  disabled={actionLoading}
                  className="w-full px-3 py-1.5 text-sm border border-border rounded-lg text-foreground hover:bg-white/[0.04] disabled:opacity-50"
                >
                  JSON 다운로드
                </button>
              </div>

              {/* Deletion Request (super_admin only) */}
              {isSuperAdmin && selectedUser.id !== currentUser?.id && (
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    삭제 요청 (PIPA)
                  </p>
                  <input
                    type="text"
                    placeholder="삭제 사유 (선택)"
                    value={deleteReason}
                    onChange={(e) => setDeleteReason(e.target.value)}
                    className="w-full px-3 py-1.5 text-sm border border-border rounded-lg bg-transparent text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
                  />
                  <button
                    onClick={handleCreateDeletion}
                    disabled={actionLoading}
                    className="w-full px-3 py-1.5 text-sm bg-red-500/10 text-red-400 rounded-lg hover:bg-red-500/20 disabled:opacity-50"
                  >
                    삭제 요청 생성 (30일 후)
                  </button>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1 text-sm border border-border rounded text-foreground disabled:opacity-50"
          >
            이전
          </button>
          <span className="text-sm text-muted-foreground">
            {page + 1} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="px-3 py-1 text-sm border border-border rounded text-foreground disabled:opacity-50"
          >
            다음
          </button>
        </div>
      )}
    </div>
  );
}

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
  last_login_at?: string;
  created_at: string;
}

const ROLE_BADGE_STYLES: Record<UserRole, string> = {
  user: "bg-gray-100 text-gray-700",
  manager: "bg-blue-100 text-blue-700",
  admin: "bg-purple-100 text-purple-700",
  super_admin: "bg-red-100 text-red-700",
};

const ROLE_LABELS: Record<UserRole, string> = {
  user: "User",
  manager: "Manager",
  admin: "Admin",
  super_admin: "Super Admin",
};

const ALL_ROLES: UserRole[] = ["user", "manager", "admin", "super_admin"];

const PAGE_SIZE = 20;

export default function AdminUsersPage() {
  const { user: currentUser } = useAuthStore();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [roleFilter, setRoleFilter] = useState("");
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);

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
        <h1 className="text-2xl font-bold text-gray-900">사용자 관리</h1>
        <p className="text-sm text-gray-500">총 {total}명</p>
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
          className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <select
          value={roleFilter}
          onChange={(e) => {
            setRoleFilter(e.target.value);
            setPage(0);
          }}
          className="px-3 py-2 border border-gray-300 rounded-lg text-sm bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">전체 역할</option>
          {ALL_ROLES.map((role) => (
            <option key={role} value={role}>
              {ROLE_LABELS[role]}
            </option>
          ))}
        </select>
      </div>

      <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-200">
            <tr>
              <th className="text-left px-4 py-3 font-medium text-gray-600">
                이메일
              </th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">
                닉네임
              </th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">
                인증
              </th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">
                역할
              </th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">
                최근 로그인
              </th>
              <th className="text-left px-4 py-3 font-medium text-gray-600">
                가입일
              </th>
              <th className="px-4 py-3" />
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={7} className="text-center py-8 text-gray-400">
                  로딩 중...
                </td>
              </tr>
            ) : users.length === 0 ? (
              <tr>
                <td colSpan={7} className="text-center py-8 text-gray-400">
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
                    className="border-b border-gray-100 hover:bg-gray-50"
                  >
                    <td className="px-4 py-3">{user.email ?? "-"}</td>
                    <td className="px-4 py-3">{user.nickname ?? "-"}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${
                          user.auth_provider === "naver"
                            ? "bg-green-100 text-green-700"
                            : "bg-blue-100 text-blue-700"
                        }`}
                      >
                        {user.auth_provider === "naver" ? "Naver" : "Email"}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${ROLE_BADGE_STYLES[user.role]}`}
                      >
                        {ROLE_LABELS[user.role]}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {user.last_login_at
                        ? new Date(user.last_login_at).toLocaleDateString(
                            "ko-KR"
                          )
                        : "-"}
                    </td>
                    <td className="px-4 py-3 text-gray-500">
                      {new Date(user.created_at).toLocaleDateString("ko-KR")}
                    </td>
                    <td className="px-4 py-3">
                      {canChange ? (
                        <select
                          value={user.role}
                          onChange={(e) =>
                            handleRoleChange(
                              user.id,
                              e.target.value as UserRole
                            )
                          }
                          className="text-xs border border-gray-300 rounded px-2 py-1 bg-white"
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
                      ) : isSelf ? (
                        <span className="text-xs text-gray-400">본인</span>
                      ) : null}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            className="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-50"
          >
            이전
          </button>
          <span className="text-sm text-gray-600">
            {page + 1} / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            className="px-3 py-1 text-sm border border-gray-300 rounded disabled:opacity-50"
          >
            다음
          </button>
        </div>
      )}
    </div>
  );
}

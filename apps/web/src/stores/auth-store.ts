import { create } from "zustand";
import { persist } from "zustand/middleware";
import { apiClient } from "@/lib/api-client";

interface UserInfo {
  id: string;
  email?: string;
  nickname?: string;
  role: "user" | "admin";
  auth_provider: "email" | "naver";
  onboarding_completed: boolean;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: UserInfo | null;
  isLoading: boolean;

  setTokens: (access: string, refresh: string) => void;
  fetchUser: () => Promise<void>;
  logout: () => void;
  refreshAccessToken: () => Promise<boolean>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isLoading: false,

      setTokens: (access, refresh) => {
        set({ accessToken: access, refreshToken: refresh });
      },

      fetchUser: async () => {
        set({ isLoading: true });
        try {
          const { data } = await apiClient.get("/v1/auth/me");
          set({ user: data, isLoading: false });
        } catch {
          set({ user: null, isLoading: false });
        }
      },

      logout: () => {
        apiClient.post("/v1/auth/logout").catch(() => {});
        set({
          accessToken: null,
          refreshToken: null,
          user: null,
        });
      },

      refreshAccessToken: async () => {
        const { refreshToken } = get();
        if (!refreshToken) return false;

        try {
          const { data } = await apiClient.post("/v1/auth/refresh", {
            refresh_token: refreshToken,
          });
          set({
            accessToken: data.tokens.access_token,
            refreshToken: data.tokens.refresh_token,
          });
          return true;
        } catch {
          get().logout();
          return false;
        }
      },
    }),
    {
      name: "colight-auth",
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    }
  )
);

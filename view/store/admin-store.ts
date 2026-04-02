import { create } from "zustand";
import { persist } from "zustand/middleware";
import {
  Admin,
  AdminDashboardStats,
  AdminActivityLog,
  AdminLoginResponse,
} from "@/types/admin";

interface AdminStore {
  // Auth
  admin: AdminLoginResponse | null;
  token: string | null;
  isAuthenticated: boolean;
  isHydrated: boolean;

  // Data
  dashboard: AdminDashboardStats | null;
  admins: Admin[];
  activityLogs: AdminActivityLog[];

  // UI State
  isLoading: boolean;
  error: string | null;

  // Actions
  setAdmin: (admin: AdminLoginResponse | null) => void;
  setToken: (token: string | null) => void;
  setDashboard: (stats: AdminDashboardStats) => void;
  setAdmins: (admins: Admin[]) => void;
  setActivityLogs: (logs: AdminActivityLog[]) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  logout: () => void;
}

export const useAdminStore = create<AdminStore>()(
  persist(
    (set) => ({
      admin: null,
      token: null,
      isAuthenticated: false,
      isHydrated: false,
      dashboard: null,
      admins: [],
      activityLogs: [],
      isLoading: false,
      error: null,

      setAdmin: (admin) =>
        set({
          admin,
          isAuthenticated: !!admin,
        }),

      setToken: (token) =>
        set({
          token,
        }),

      setDashboard: (dashboard) =>
        set({
          dashboard,
        }),

      setAdmins: (admins) =>
        set({
          admins,
        }),

      setActivityLogs: (activityLogs) =>
        set({
          activityLogs,
        }),

      setLoading: (isLoading) =>
        set({
          isLoading,
        }),

      setError: (error) =>
        set({
          error,
        }),

      logout: () =>
        set({
          admin: null,
          token: null,
          isAuthenticated: false,
          dashboard: null,
          admins: [],
          activityLogs: [],
          error: null,
          isHydrated: true,
        }),
    }),
    {
      name: "admin-store",
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.isHydrated = true;
        }
      },
      partialize: (state) => ({
        admin: state.admin,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    },
  ),
);

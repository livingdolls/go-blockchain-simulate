import { api } from "@/lib/axios";
import {
  AdminLoginRequest,
  AdminLoginResponse,
  Admin,
  AdminActivityLog,
  AdminDashboardStats,
  CreateAdminRequest,
  UpdateAdminRoleRequest,
  UpdateAdminStatusRequest,
} from "@/types/admin";

export const adminRepository = {
  // Authentication
  login: async (data: AdminLoginRequest): Promise<AdminLoginResponse> => {
    const response = await api.post<{ data: AdminLoginResponse }>(
      "/admin/auth/login",
      data,
    );
    return response.data.data;
  },

  logout: async (): Promise<void> => {
    await api.post("/admin/auth/logout");
  },

  // Dashboard
  getDashboard: async (): Promise<AdminDashboardStats> => {
    const response = await api.get<{ data: AdminDashboardStats }>(
      "/admin/dashboard",
    );
    return response.data.data;
  },

  // Admin Management
  getAdmins: async (
    limit: number = 10,
    offset: number = 0,
  ): Promise<Admin[]> => {
    const response = await api.get<{ data: Admin[] }>("/admin/admins", {
      params: { limit, offset },
    });
    return response.data.data;
  },

  createAdmin: async (data: CreateAdminRequest): Promise<Admin> => {
    const response = await api.post<{ data: Admin }>("/admin/admins", data);
    return response.data.data;
  },

  updateAdminRole: async (
    id: number,
    data: UpdateAdminRoleRequest,
  ): Promise<Admin> => {
    const response = await api.put<{ data: Admin }>(
      `/admin/admins/${id}/role`,
      data,
    );
    return response.data.data;
  },

  updateAdminStatus: async (
    id: number,
    data: UpdateAdminStatusRequest,
  ): Promise<Admin> => {
    const response = await api.put<{ data: Admin }>(
      `/admin/admins/${id}/status`,
      data,
    );
    return response.data.data;
  },

  deleteAdmin: async (id: number): Promise<{ message: string }> => {
    const response = await api.delete<{ message: string }>(
      `/admin/admins/${id}`,
    );
    return response.data;
  },

  // Activity Logs
  getActivityLogs: async (
    adminId?: number,
    action?: string,
    limit: number = 20,
    offset: number = 0,
  ): Promise<AdminActivityLog[]> => {
    const response = await api.get<{ data: AdminActivityLog[] }>(
      "/admin/activity-logs",
      {
        params: { admin_id: adminId, action, limit, offset },
      },
    );
    return response.data.data;
  },

  getRecentActivityLogs: async (
    days: number = 7,
    limit: number = 50,
  ): Promise<AdminActivityLog[]> => {
    const response = await api.get<{ data: AdminActivityLog[] }>(
      "/admin/activity-logs/recent",
      {
        params: { days, limit },
      },
    );
    return response.data.data;
  },
};

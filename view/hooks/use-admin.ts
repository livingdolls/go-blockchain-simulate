import { useMutation, useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useAdminStore } from "@/store/admin-store";
import { adminRepository } from "@/repository/admin";
import {
  AdminLoginRequest,
  CreateAdminRequest,
  UpdateAdminRoleRequest,
  UpdateAdminStatusRequest,
} from "@/types/admin";
import { toast } from "sonner";
import axios from "axios";

// Login Hook
export const useAdminLogin = () => {
  const router = useRouter();
  const { setAdmin, setToken } = useAdminStore();

  return useMutation({
    mutationFn: (data: AdminLoginRequest) => adminRepository.login(data),
    onSuccess: (data) => {
      setAdmin(data);
      setToken(data.token);
      toast.success("Login berhasil!");
      router.push("/admin/dashboard");
    },
    onError: (error: unknown) => {
      let message = "Login gagal";

      if (axios.isAxiosError(error)) {
        message = error.response?.data?.message ?? message;
      }

      toast.error(message);
    },
  });
};

// Logout Hook
export const useAdminLogout = () => {
  const router = useRouter();
  const { logout } = useAdminStore();

  return useMutation({
    mutationFn: () => adminRepository.logout(),
    onSuccess: () => {
      logout();
      toast.success("Logout berhasil!");
      router.push("/admin/login");
    },
    onError: (error: unknown) => {
      let message = "Logout gagal";

      if (axios.isAxiosError(error)) {
        message = error.response?.data?.message ?? message;
      }

      toast.error(message);
    },
  });
};

// Dashboard Hook
export const useAdminDashboard = () => {
  const { setDashboard, setLoading, setError } = useAdminStore();

  return useQuery({
    queryKey: ["admin-dashboard"],
    queryFn: async () => {
      try {
        setLoading(true);
        const data = await adminRepository.getDashboard();
        setDashboard(data);
        return data;
      } catch (error: unknown) {
        let message = "Gagal mengambil dashboard";

        if (axios.isAxiosError(error)) {
          message = error.response?.data?.message ?? message;
        }

        setError(message);
        throw error;
      } finally {
        setLoading(false);
      }
    },
    refetchInterval: 30000, // Refetch setiap 30 detik
  });
};

// Get Admins Hook
export const useGetAdmins = (limit: number = 10, offset: number = 0) => {
  const { setAdmins, setLoading, setError } = useAdminStore();

  return useQuery({
    queryKey: ["admin-list", limit, offset],
    queryFn: async () => {
      try {
        setLoading(true);
        const data = await adminRepository.getAdmins(limit, offset);
        setAdmins(data);
        return data;
      } catch (error: unknown) {
        let message = "Gagal mengambil daftar admin";

        if (axios.isAxiosError(error)) {
          message = error.response?.data?.message ?? message;
        }

        setError(message);
        throw error;
      } finally {
        setLoading(false);
      }
    },
  });
};

// Create Admin Hook
export const useCreateAdmin = () => {
  return useMutation({
    mutationFn: (data: CreateAdminRequest) => adminRepository.createAdmin(data),
    onSuccess: () => {
      toast.success("Admin berhasil dibuat!");
    },
    onError: (error: unknown) => {
      let message = "Gagal membuat admin";

      if (axios.isAxiosError(error)) {
        message = error.response?.data?.message ?? message;
      }

      toast.error(message);
    },
  });
};

// Update Admin Role Hook
export const useUpdateAdminRole = () => {
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateAdminRoleRequest }) =>
      adminRepository.updateAdminRole(id, data),
    onSuccess: () => {
      toast.success("Role admin berhasil diperbarui!");
    },
    onError: (error: unknown) => {
      let message = "Gagal memperbarui role";

      if (axios.isAxiosError(error)) {
        message = error.response?.data?.message ?? message;
      }

      toast.error(message);
    },
  });
};

// Update Admin Status Hook
export const useUpdateAdminStatus = () => {
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: UpdateAdminStatusRequest;
    }) => adminRepository.updateAdminStatus(id, data),
    onSuccess: () => {
      toast.success("Status admin berhasil diperbarui!");
    },
    onError: (error: unknown) => {
      let message = "Gagal memperbarui status";

      if (axios.isAxiosError(error)) {
        message = error.response?.data?.message ?? message;
      }

      toast.error(message);
    },
  });
};

// Delete Admin Hook
export const useDeleteAdmin = () => {
  return useMutation({
    mutationFn: (id: number) => adminRepository.deleteAdmin(id),
    onSuccess: () => {
      toast.success("Admin berhasil dihapus!");
    },
    onError: (error: unknown) => {
      let message = "Gagal menghapus admin";

      if (axios.isAxiosError(error)) {
        message = error.response?.data?.message ?? message;
      }

      toast.error(message);
    },
  });
};

// Activity Logs Hook
export const useActivityLogs = (adminId?: number, action?: string) => {
  const { setActivityLogs, setLoading, setError } = useAdminStore();

  return useQuery({
    queryKey: ["activity-logs", adminId, action],
    queryFn: async () => {
      try {
        setLoading(true);
        const data = await adminRepository.getActivityLogs(
          adminId,
          action,
          50,
          0,
        );
        setActivityLogs(data);
        return data;
      } catch (error: unknown) {
        let message = "Gagal mengambil activity logs";

        if (axios.isAxiosError(error)) {
          message = error.response?.data?.message ?? message;
        }

        setError(message);
        throw error;
      } finally {
        setLoading(false);
      }
    },
  });
};

// Recent Activity Logs Hook
export const useRecentActivityLogs = (days: number = 7) => {
  const { setActivityLogs, setLoading } = useAdminStore();

  return useQuery({
    queryKey: ["recent-activity-logs", days],
    queryFn: async () => {
      try {
        setLoading(true);
        const data = await adminRepository.getRecentActivityLogs(days, 50);
        setActivityLogs(data);
        return data;
      } catch (error: unknown) {
        let message = "Gagal mengambil recent activity logs";

        if (axios.isAxiosError(error)) {
          // eslint-disable-next-line @typescript-eslint/no-unused-vars
          message = error.response?.data?.message ?? message;
        }

        throw error;
      } finally {
        setLoading(false);
      }
    },
    refetchInterval: 10000, // Refetch setiap 10 detik
  });
};

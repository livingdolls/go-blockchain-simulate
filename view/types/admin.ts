export interface AdminLoginRequest {
  username: string;
  password: string;
}

export interface AdminLoginResponse {
  id: number;
  user_id: number;
  username: string;
  role: "admin" | "moderator" | "support";
  token: string;
}

export interface Admin {
  id: number;
  user_id: number;
  username: string;
  address: string;
  role: "admin" | "moderator" | "support";
  permissions: string[];
  status: "active" | "inactive" | "suspended";
  last_login_at: string;
  created_at: string;
}

export interface AdminActivityLog {
  id: number;
  admin_id: number;
  action: "create" | "update" | "delete" | "read";
  target_entity: string;
  target_id: number;
  target_name: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  old_values?: Record<string, any>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  new_values?: Record<string, any>;
  changes_summary: string;
  ip_address: string;
  user_agent: string;
  status: "success" | "failed";
  error_message?: string;
  created_at: string;
}

export interface AdminDashboardStats {
  total_users: number;
  total_transactions: number;
  total_blocks: number;
  total_admins: number;
  active_users: number;
  suspended_admins: number;
  recent_activity_count: number;
  total_volume: number;
}

export interface CreateAdminRequest {
  user_id: number;
  role: "admin" | "moderator" | "support";
  permissions: string[];
}

export interface UpdateAdminRoleRequest {
  role: "admin" | "moderator" | "support";
  permissions: string[];
}

export interface UpdateAdminStatusRequest {
  status: "active" | "inactive" | "suspended";
}

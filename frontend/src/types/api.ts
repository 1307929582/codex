export interface User {
  id: string;
  email: string;
  username?: string;
  avatar_url?: string;
  oauth_provider?: string;
  oauth_id?: string;
  balance: number;
  status: string;
  role: string; // user, admin, super_admin
  created_at: string;
}

export interface APIKey {
  id: number;
  user_id: string;
  key_prefix: string;
  name: string;
  quota_limit: number | null;
  total_usage: number;
  status: string;
  created_at: string;
  last_used_at: string | null;
}

export interface UsageLog {
  request_id: string;
  user_id: string;
  api_key_id: number;
  model: string;
  input_tokens: number;
  output_tokens: number;
  cached_tokens: number;
  total_tokens: number;
  cost: number;
  latency_ms: number;
  status_code: number;
  created_at: string;
}

export interface AdminUsageLog {
  request_id: string;
  user_id: string;
  username?: string;
  linuxdo_id?: string;
  api_key_id: number;
  model: string;
  input_tokens: number;
  output_tokens: number;
  cached_tokens: number;
  total_tokens: number;
  cost: number;
  latency_ms: number;
  status_code: number;
  created_at: string;
}

export interface UsageStats {
  today_cost: number;
  month_cost: number;
  total_cost: number;
}

export interface Transaction {
  id: string;
  user_id: string;
  amount: number;
  type: string;
  description: string;
  created_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface CreateKeyRequest {
  name: string;
  quota_limit?: number;
}

export interface CreateKeyResponse {
  id: number;
  key: string;
  name: string;
}

// Admin types
export interface AdminUser extends User {
  api_key_count?: number;
  total_cost?: number;
  total_tokens?: number;
  active_package?: UserPackage;
}

export interface SystemSettings {
  id: number;
  announcement: string;
  default_balance: number;
  min_recharge_amount: number;
  email_registration_enabled: boolean;
  linuxdo_registration_enabled: boolean;
  openai_api_key: string;
  openai_base_url: string;
  linuxdo_client_id: string;
  linuxdo_client_secret: string;
  linuxdo_enabled: boolean;
  credit_enabled: boolean;
  credit_pid: string;
  credit_key: string;
  credit_notify_url: string;
  credit_return_url: string;
  created_at: string;
  updated_at: string;
}

export interface Package {
  id: number;
  name: string;
  description: string;
  price: number;
  duration_days: number;
  daily_limit: number;
  status: string;
  sort_order: number;
  stock: number;        // -1 means unlimited
  sold_count: number;   // Number of packages sold
  created_at: string;
  updated_at: string;
}

export interface UserPackage {
  id: string;
  user_id: string;
  package_id: number;
  package_name: string;
  package_price: number;
  duration_days: number;
  daily_limit: number;
  start_date: string;
  end_date: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface DailyUsage {
  date: string;
  used_amount: number;
  package?: UserPackage;
  daily_limit?: number;
  remaining?: number;
}

export interface AdminLog {
  id: number;
  admin_id: string;
  action: string;
  target: string;
  details: string;
  ip_address: string;
  created_at: string;
}

export interface AdminStats {
  total_users: number;
  active_users: number;
  total_tokens: number;
  total_cost: number;
  total_api_keys: number;
  today_requests: number;
  today_revenue: number;
}

export interface HourlyUsage {
  hour: string;
  cost: number;
}

export interface PaginationResponse<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

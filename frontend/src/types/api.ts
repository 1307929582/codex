export interface User {
  id: string;
  email: string;
  balance: number;
  status: string;
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

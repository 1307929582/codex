import { api } from './client';
import type {
  AdminUser,
  AdminStats,
  SystemSettings,
  AdminLog,
  PaginationResponse,
  HourlyUsage,
} from '@/types/api';

export const adminApi = {
  // User Management
  getUsers: async (params?: {
    page?: number;
    page_size?: number;
    search?: string;
    status?: string;
  }) => {
    const response = await api.get<{
      users: AdminUser[];
      pagination: {
        page: number;
        page_size: number;
        total: number;
        total_pages: number;
      };
    }>('/api/admin/users', { params });
    return response.data;
  },

  getUser: async (id: string) => {
    const response = await api.get<{
      user: AdminUser;
      api_key_count: number;
      total_cost: number;
      total_tokens: number;
    }>(`/api/admin/users/${id}`);
    return response.data;
  },

  updateBalance: async (id: string, amount: number, description: string) => {
    const response = await api.put(`/api/admin/users/${id}/balance`, {
      amount,
      description,
    });
    return response.data;
  },

  updateUserStatus: async (id: string, status: 'active' | 'suspended' | 'banned') => {
    const response = await api.put(`/api/admin/users/${id}/status`, {
      status,
    });
    return response.data;
  },

  // System Settings
  getSettings: async () => {
    const response = await api.get<SystemSettings>('/api/admin/settings');
    return response.data;
  },

  updateSettings: async (settings: Partial<SystemSettings>) => {
    const response = await api.put('/api/admin/settings', settings);
    return response.data;
  },

  // Statistics
  getOverview: async () => {
    const response = await api.get<AdminStats>('/api/admin/stats/overview');
    return response.data;
  },

  getUsageChart: async (range: '24h' | '7d' | '30d' = '24h') => {
    const response = await api.get<Array<{ label: string; cost: number }>>('/api/admin/stats/usage-chart', {
      params: { range },
    });
    return response.data;
  },

  // Logs
  getLogs: async (params?: { page?: number; page_size?: number }) => {
    const response = await api.get<{
      logs: AdminLog[];
      pagination: {
        page: number;
        page_size: number;
        total: number;
        total_pages: number;
      };
    }>('/api/admin/logs', { params });
    return response.data;
  },

  // Codex Upstreams
  getCodexUpstreams: async () => {
    const response = await api.get<{
      upstreams: Array<{
        id: number;
        name: string;
        base_url: string;
        api_key: string;
        priority: number;
        status: 'active' | 'disabled' | 'unhealthy';
        weight: number;
        max_retries: number;
        timeout: number;
        health_check: string;
        last_checked: string | null;
        created_at: string;
        updated_at: string;
      }>;
    }>('/api/admin/codex/upstreams');
    return response.data;
  },

  getCodexUpstream: async (id: number) => {
    const response = await api.get(`/api/admin/codex/upstreams/${id}`);
    return response.data;
  },

  createCodexUpstream: async (data: {
    name: string;
    base_url: string;
    api_key: string;
    priority: number;
    status: string;
    weight: number;
    max_retries: number;
    timeout: number;
  }) => {
    const response = await api.post('/api/admin/codex/upstreams', data);
    return response.data;
  },

  updateCodexUpstream: async (id: number, data: {
    name: string;
    base_url: string;
    api_key: string;
    priority: number;
    status: string;
    weight: number;
    max_retries: number;
    timeout: number;
  }) => {
    const response = await api.put(`/api/admin/codex/upstreams/${id}`, data);
    return response.data;
  },

  deleteCodexUpstream: async (id: number) => {
    const response = await api.delete(`/api/admin/codex/upstreams/${id}`);
    return response.data;
  },

  updateCodexUpstreamStatus: async (id: number, status: 'active' | 'disabled' | 'unhealthy') => {
    const response = await api.put(`/api/admin/codex/upstreams/${id}/status`, { status });
    return response.data;
  },

  getUpstreamHealth: async () => {
    const response = await api.get<{
      upstreams: Array<{
        id: number;
        name: string;
        status: string;
        failure_count: number;
        last_checked: string;
      }>;
    }>('/api/admin/codex/upstreams/health');
    return response.data;
  },

  triggerHealthCheck: async () => {
    const response = await api.post('/api/admin/codex/upstreams/health/check');
    return response.data;
  },
};

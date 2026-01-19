import { api } from './client';
import type {
  AdminUser,
  AdminStats,
  SystemSettings,
  AdminLog,
  PaginationResponse,
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
};

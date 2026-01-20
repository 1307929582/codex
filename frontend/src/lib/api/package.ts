import { api } from './client';
import type { Package, UserPackage, DailyUsage } from '@/types/api';

export const packageApi = {
  // Admin APIs
  adminList: async () => {
    const response = await api.get<{ packages: Package[] }>('/api/admin/packages');
    return response.data;
  },

  adminCreate: async (data: {
    name: string;
    description: string;
    price: number;
    duration_days: number;
    daily_limit: number;
    sort_order?: number;
    stock?: number;
  }) => {
    const response = await api.post<Package>('/api/admin/packages', data);
    return response.data;
  },

  adminUpdate: async (id: number, data: {
    name?: string;
    description?: string;
    price?: number;
    duration_days?: number;
    daily_limit?: number;
    sort_order?: number;
    stock?: number;
  }) => {
    const response = await api.put<Package>(`/api/admin/packages/${id}`, data);
    return response.data;
  },

  adminDelete: async (id: number) => {
    const response = await api.delete(`/api/admin/packages/${id}`);
    return response.data;
  },

  adminUpdateStatus: async (id: number, status: 'active' | 'inactive') => {
    const response = await api.put(`/api/admin/packages/${id}/status`, { status });
    return response.data;
  },

  // User APIs
  list: async () => {
    const response = await api.get<{ packages: Package[] }>('/api/packages');
    return response.data;
  },

  purchase: async (id: number) => {
    const response = await api.post<{
      order_no: string;
      amount: number;
      payment_url: string;
      params: Record<string, string>;
    }>(`/api/packages/${id}/purchase`);
    return response.data;
  },

  getUserPackages: async (params?: { page?: number; page_size?: number }) => {
    const response = await api.get<{
      packages: UserPackage[];
      pagination: {
        page: number;
        page_size: number;
        total: number;
        total_pages: number;
      };
    }>('/api/user/packages', { params });
    return response.data;
  },

  getDailyUsage: async () => {
    const response = await api.get<DailyUsage>('/api/user/daily-usage');
    return response.data;
  },
};

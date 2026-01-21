import { api } from './client';
import type { Coupon } from '@/types/api';

export const couponApi = {
  list: async (params?: { page?: number; page_size?: number; search?: string; status?: string }) => {
    const response = await api.get<{
      coupons: Coupon[];
      pagination: {
        page: number;
        page_size: number;
        total: number;
        total_pages: number;
      };
    }>('/api/admin/coupons', { params });
    return response.data;
  },

  create: async (data: {
    code: string;
    type: 'fixed' | 'percent';
    value: number;
    max_uses?: number;
    min_amount?: number;
    starts_at?: string;
    ends_at?: string;
    status?: string;
  }) => {
    const response = await api.post<Coupon>('/api/admin/coupons', data);
    return response.data;
  },

  update: async (id: number, data: {
    code?: string;
    type?: 'fixed' | 'percent';
    value?: number;
    max_uses?: number;
    min_amount?: number;
    starts_at?: string;
    ends_at?: string;
    status?: string;
  }) => {
    const response = await api.put<Coupon>(`/api/admin/coupons/${id}`, data);
    return response.data;
  },
};

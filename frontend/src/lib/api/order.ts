import api from './client';

export interface PaymentOrder {
  id: string;
  order_no: string;
  user_id: string;
  user_email: string;
  username?: string;
  package_id: number | null;
  amount: number;
  status: string;
  payment_method: string;
  trade_no: string;
  created_at: string;
  paid_at: string | null;
}

export interface OrderStats {
  total_orders: number;
  pending_orders: number;
  paid_orders: number;
  failed_orders: number;
  total_revenue: number;
  today_revenue: number;
  month_revenue: number;
}

export interface UserPackageAdmin {
  id: string;
  user_id: string;
  user_email: string;
  package_id: number;
  package_name: string;
  package_price: number;
  duration_days: number;
  daily_limit: number;
  start_date: string;
  end_date: string;
  status: string;
  created_at: string;
}

export const orderApi = {
  // List orders
  list: async (params?: { page?: number; page_size?: number; status?: string; user_id?: string }) => {
    const response = await api.get<{
      orders: PaymentOrder[];
      pagination: {
        page: number;
        page_size: number;
        total: number;
        total_pages: number;
      };
    }>('/api/admin/orders', { params });
    return response.data;
  },

  // Get order statistics
  getStats: async () => {
    const response = await api.get<OrderStats>('/api/admin/orders/stats');
    return response.data;
  },

  // List user packages
  listUserPackages: async (params?: { page?: number; page_size?: number; status?: string; user_id?: string }) => {
    const response = await api.get<{
      user_packages: UserPackageAdmin[];
      pagination: {
        page: number;
        page_size: number;
        total: number;
        total_pages: number;
      };
    }>('/api/admin/user-packages', { params });
    return response.data;
  },
};

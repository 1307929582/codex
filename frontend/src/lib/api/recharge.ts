import { api } from './client';

export const rechargeApi = {
  // Create recharge order
  createOrder: async (amount: number) => {
    const response = await api.post<{
      order_no: string;
      amount: number;
      payment_url: string;
      params: Record<string, string>;
    }>('/api/recharge', { amount });
    return response.data;
  },
};

import { api } from './client';

export interface ModelPricing {
  id: number;
  model_name: string;
  input_price_per_1k: number;
  output_price_per_1k: number;
  cache_read_price_per_1k: number;
  markup_multiplier: number;
  effective_from: string;
}

export const pricingApi = {
  // List all model pricing
  list: async () => {
    const response = await api.get<{ pricing: ModelPricing[] }>('/api/admin/pricing');
    return response.data;
  },

  // Update a model's pricing
  update: async (id: number, data: Partial<ModelPricing>) => {
    const response = await api.put(`/api/admin/pricing/${id}`, data);
    return response.data;
  },

  // Batch update markup multiplier for all models
  batchUpdateMarkup: async (markupMultiplier: number) => {
    const response = await api.post('/api/admin/pricing/batch-update-markup', {
      markup_multiplier: markupMultiplier,
    });
    return response.data;
  },

  // Reset pricing to default values
  reset: async () => {
    const response = await api.post('/api/admin/pricing/reset');
    return response.data;
  },
};

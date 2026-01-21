'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { pricingApi, ModelPricing } from '@/lib/api/pricing';
import { Edit, RefreshCw, Percent } from 'lucide-react';

export default function AdminPricingPage() {
  const queryClient = useQueryClient();
  const [editingPricing, setEditingPricing] = useState<ModelPricing | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showBatchModal, setShowBatchModal] = useState(false);
  const [batchMarkup, setBatchMarkup] = useState('1.5');

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'pricing'],
    queryFn: () => pricingApi.list(),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<ModelPricing> }) =>
      pricingApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'pricing'] });
      setShowEditModal(false);
      setEditingPricing(null);
      alert('定价更新成功！');
    },
    onError: (error: any) => {
      alert(`更新失败: ${error?.response?.data?.error || error?.message}`);
    },
  });

  const batchUpdateMutation = useMutation({
    mutationFn: pricingApi.batchUpdateMarkup,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'pricing'] });
      setShowBatchModal(false);
      alert(`批量更新成功！已更新 ${data.updated_count} 个模型的比例`);
    },
    onError: (error: any) => {
      alert(`批量更新失败: ${error?.response?.data?.error || error?.message}`);
    },
  });

  const resetMutation = useMutation({
    mutationFn: pricingApi.reset,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'pricing'] });
      alert('定价已重置为默认值！');
    },
    onError: (error: any) => {
      alert(`重置失败: ${error?.response?.data?.error || error?.message}`);
    },
  });

  const handleEdit = (pricing: ModelPricing) => {
    setEditingPricing(pricing);
    setShowEditModal(true);
  };

  const handleUpdate = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editingPricing) return;

    const formData = new FormData(e.currentTarget);
    // Convert from $/1M (user input) to $/1K (backend storage)
    const data = {
      input_price_per_1k: parseFloat(formData.get('input_price_per_1k') as string) / 1000,
      output_price_per_1k: parseFloat(formData.get('output_price_per_1k') as string) / 1000,
      cache_read_price_per_1k: parseFloat(formData.get('cache_read_price_per_1k') as string) / 1000,
      cache_creation_price_per_1k: parseFloat(formData.get('cache_creation_price_per_1k') as string) / 1000,
      markup_multiplier: parseFloat(formData.get('markup_multiplier') as string),
    };

    updateMutation.mutate({ id: editingPricing.id, data });
  };

  const handleBatchUpdate = () => {
    const markup = parseFloat(batchMarkup);
    if (isNaN(markup) || markup <= 0) {
      alert('请输入有效的比例值（大于0）');
      return;
    }

    if (confirm(`确定要将所有模型的价格比例设置为 ${markup}x 吗？`)) {
      batchUpdateMutation.mutate(markup);
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl space-y-6">
      <div className="flex items-center justify-between border-b border-zinc-200 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">定价管理</h1>
          <p className="text-sm text-zinc-500">管理模型定价和价格比例</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowBatchModal(true)}
            className="inline-flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
          >
            <Percent className="h-4 w-4" />
            批量设置比例
          </button>
          <button
            onClick={() => {
              if (confirm('确定要重置所有定价为默认值吗？此操作不可恢复。')) {
                resetMutation.mutate();
              }
            }}
            disabled={resetMutation.isPending}
            className="inline-flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50 disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${resetMutation.isPending ? 'animate-spin' : ''}`} />
            重置定价
          </button>
        </div>
      </div>

      <div className="rounded-xl border border-zinc-200 bg-white shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-zinc-50 border-b border-zinc-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  模型名称
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  输入价格 ($/1M)
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  输出价格 ($/1M)
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  缓存读取 ($/1M)
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  缓存创建 ($/1M)
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  价格比例
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-200">
              {data?.pricing.map((pricing) => (
                <tr key={pricing.id} className="hover:bg-zinc-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-zinc-900">
                    {pricing.model_name}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-zinc-600">
                    ${(pricing.input_price_per_1k * 1000).toFixed(2)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-zinc-600">
                    ${(pricing.output_price_per_1k * 1000).toFixed(2)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-zinc-600">
                    ${(pricing.cache_read_price_per_1k * 1000).toFixed(2)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right text-zinc-600">
                    ${(((pricing.cache_creation_price_per_1k ?? pricing.cache_read_price_per_1k) || 0) * 1000).toFixed(2)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right">
                    <span className="inline-flex items-center rounded-full bg-blue-50 px-2.5 py-0.5 text-xs font-medium text-blue-700 ring-1 ring-blue-600/20">
                      {pricing.markup_multiplier}x
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      onClick={() => handleEdit(pricing)}
                      className="inline-flex items-center gap-1 rounded-lg border border-zinc-200 px-3 py-1.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
                    >
                      <Edit className="h-3.5 w-3.5" />
                      编辑
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Edit Modal */}
      {showEditModal && editingPricing && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-zinc-900">
              编辑定价 - {editingPricing.model_name}
            </h3>
            <form onSubmit={handleUpdate} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  输入价格 ($/1M tokens)
                </label>
                <input
                  type="number"
                  name="input_price_per_1k"
                  step="0.01"
                  defaultValue={(editingPricing.input_price_per_1k * 1000).toFixed(2)}
                  required
                  className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  显示为 $/1M，实际存储为 $/1K
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  输出价格 ($/1M tokens)
                </label>
                <input
                  type="number"
                  name="output_price_per_1k"
                  step="0.01"
                  defaultValue={(editingPricing.output_price_per_1k * 1000).toFixed(2)}
                  required
                  className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  显示为 $/1M，实际存储为 $/1K
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  缓存读取价格 ($/1M tokens)
                </label>
                <input
                  type="number"
                  name="cache_read_price_per_1k"
                  step="0.01"
                  defaultValue={(editingPricing.cache_read_price_per_1k * 1000).toFixed(2)}
                  required
                  className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  显示为 $/1M，实际存储为 $/1K
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  缓存创建价格 ($/1M tokens)
                </label>
                <input
                  type="number"
                  name="cache_creation_price_per_1k"
                  step="0.01"
                  defaultValue={(((editingPricing.cache_creation_price_per_1k ?? editingPricing.cache_read_price_per_1k) || 0) * 1000).toFixed(2)}
                  required
                  className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  显示为 $/1M，实际存储为 $/1K
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  价格比例 (倍数)
                </label>
                <input
                  type="number"
                  name="markup_multiplier"
                  step="0.01"
                  defaultValue={editingPricing.markup_multiplier}
                  required
                  className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  实际收费 = 成本 × 价格比例
                </p>
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowEditModal(false);
                    setEditingPricing(null);
                  }}
                  className="flex-1 rounded-lg border border-zinc-200 px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={updateMutation.isPending}
                  className="flex-1 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50"
                >
                  {updateMutation.isPending ? '保存中...' : '保存'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Batch Update Modal */}
      {showBatchModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-semibold text-zinc-900">
              批量设置价格比例
            </h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  价格比例 (倍数)
                </label>
                <input
                  type="number"
                  step="0.01"
                  value={batchMarkup}
                  onChange={(e) => setBatchMarkup(e.target.value)}
                  placeholder="例如: 1.5"
                  className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  将应用到所有模型。实际收费 = 成本 × 价格比例
                </p>
              </div>

              <div className="rounded-lg bg-amber-50 p-4 text-sm text-amber-700">
                <p className="font-medium mb-1">注意</p>
                <p className="text-xs">
                  此操作将更新所有模型的价格比例，请谨慎操作。
                </p>
              </div>

              <div className="flex gap-3">
                <button
                  onClick={() => setShowBatchModal(false)}
                  className="flex-1 rounded-lg border border-zinc-200 px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
                >
                  取消
                </button>
                <button
                  onClick={handleBatchUpdate}
                  disabled={batchUpdateMutation.isPending}
                  className="flex-1 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50"
                >
                  {batchUpdateMutation.isPending ? '更新中...' : '确认更新'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

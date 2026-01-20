'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { packageApi } from '@/lib/api/package';
import { Plus, Edit, Trash2, Power, PowerOff } from 'lucide-react';

export default function AdminPackagesPage() {
  const queryClient = useQueryClient();
  const [showModal, setShowModal] = useState(false);
  const [editingPackage, setEditingPackage] = useState<any>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'packages'],
    queryFn: () => packageApi.adminList(),
  });

  const createMutation = useMutation({
    mutationFn: packageApi.adminCreate,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'packages'] });
      setShowModal(false);
      setEditingPackage(null);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      packageApi.adminUpdate(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'packages'] });
      setShowModal(false);
      setEditingPackage(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: packageApi.adminDelete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'packages'] });
    },
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: 'active' | 'inactive' }) =>
      packageApi.adminUpdateStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'packages'] });
    },
  });

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const data = {
      name: formData.get('name') as string,
      description: formData.get('description') as string,
      price: parseFloat(formData.get('price') as string),
      duration_days: parseInt(formData.get('duration_days') as string),
      daily_limit: parseFloat(formData.get('daily_limit') as string),
      sort_order: parseInt(formData.get('sort_order') as string) || 0,
      stock: parseInt(formData.get('stock') as string) || -1,
    };

    if (editingPackage) {
      updateMutation.mutate({ id: editingPackage.id, data });
    } else {
      createMutation.mutate(data);
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
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">套餐管理</h1>
          <p className="text-sm text-zinc-500">管理平台套餐配置</p>
        </div>
        <button
          onClick={() => {
            setEditingPackage(null);
            setShowModal(true);
          }}
          className="inline-flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800"
        >
          <Plus className="h-4 w-4" />
          创建套餐
        </button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {data?.packages.map((pkg) => (
          <div
            key={pkg.id}
            className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm transition-all hover:shadow-md"
          >
            <div className="mb-4 flex items-start justify-between">
              <div>
                <h3 className="text-lg font-bold text-zinc-900">{pkg.name}</h3>
                <p className="mt-1 text-sm text-zinc-500">{pkg.description}</p>
              </div>
              <span
                className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                  pkg.status === 'active'
                    ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20'
                    : 'bg-zinc-100 text-zinc-600'
                }`}
              >
                {pkg.status === 'active' ? '启用' : '禁用'}
              </span>
            </div>

            <div className="space-y-2 border-t border-zinc-100 pt-4">
              <div className="flex justify-between text-sm">
                <span className="text-zinc-500">价格</span>
                <span className="font-medium text-zinc-900">${pkg.price.toFixed(2)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-zinc-500">有效期</span>
                <span className="font-medium text-zinc-900">{pkg.duration_days}天</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-zinc-500">每日限额</span>
                <span className="font-medium text-zinc-900">${pkg.daily_limit.toFixed(2)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-zinc-500">库存</span>
                <span className={`font-medium ${pkg.stock === -1 ? 'text-green-600' : pkg.stock > 0 ? 'text-zinc-900' : 'text-red-600'}`}>
                  {pkg.stock === -1 ? '无限' : pkg.stock > 0 ? `${pkg.stock}份` : '售罄'}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-zinc-500">已售</span>
                <span className="font-medium text-zinc-900">{pkg.sold_count || 0}份</span>
              </div>
            </div>

            <div className="mt-4 flex gap-2 border-t border-zinc-100 pt-4">
              <button
                onClick={() => {
                  setEditingPackage(pkg);
                  setShowModal(true);
                }}
                className="flex-1 rounded-lg border border-zinc-200 px-3 py-1.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
              >
                <Edit className="mx-auto h-4 w-4" />
              </button>
              <button
                onClick={() => {
                  if (confirm(`确定要${pkg.status === 'active' ? '禁用' : '启用'}此套餐吗？`)) {
                    updateStatusMutation.mutate({
                      id: pkg.id,
                      status: pkg.status === 'active' ? 'inactive' : 'active',
                    });
                  }
                }}
                className="flex-1 rounded-lg border border-zinc-200 px-3 py-1.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
              >
                {pkg.status === 'active' ? (
                  <PowerOff className="mx-auto h-4 w-4" />
                ) : (
                  <Power className="mx-auto h-4 w-4" />
                )}
              </button>
              <button
                onClick={() => {
                  if (confirm('确定要删除此套餐吗？此操作不可恢复。')) {
                    deleteMutation.mutate(pkg.id);
                  }
                }}
                className="flex-1 rounded-lg border border-red-200 px-3 py-1.5 text-sm font-medium text-red-600 transition-colors hover:bg-red-50"
              >
                <Trash2 className="mx-auto h-4 w-4" />
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create/Edit Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
          <div className="w-full max-w-md rounded-lg bg-white p-6">
            <h3 className="mb-4 text-xl font-bold text-zinc-900">
              {editingPackage ? '编辑套餐' : '创建套餐'}
            </h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700">套餐名称</label>
                <input
                  type="text"
                  name="name"
                  defaultValue={editingPackage?.name}
                  required
                  className="mt-1 w-full rounded-lg border border-zinc-300 px-4 py-2 focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700">描述</label>
                <textarea
                  name="description"
                  defaultValue={editingPackage?.description}
                  rows={3}
                  className="mt-1 w-full rounded-lg border border-zinc-300 px-4 py-2 focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-zinc-700">价格 ($)</label>
                  <input
                    type="number"
                    name="price"
                    step="0.01"
                    defaultValue={editingPackage?.price}
                    required
                    className="mt-1 w-full rounded-lg border border-zinc-300 px-4 py-2 focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-zinc-700">有效期 (天)</label>
                  <input
                    type="number"
                    name="duration_days"
                    defaultValue={editingPackage?.duration_days}
                    required
                    className="mt-1 w-full rounded-lg border border-zinc-300 px-4 py-2 focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-zinc-700">每日限额 ($)</label>
                  <input
                    type="number"
                    name="daily_limit"
                    step="0.01"
                    defaultValue={editingPackage?.daily_limit}
                    required
                    className="mt-1 w-full rounded-lg border border-zinc-300 px-4 py-2 focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-zinc-700">库存 (-1=无限)</label>
                  <input
                    type="number"
                    name="stock"
                    defaultValue={editingPackage?.stock ?? -1}
                    className="mt-1 w-full rounded-lg border border-zinc-300 px-4 py-2 focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-zinc-700">排序</label>
                  <input
                    type="number"
                    name="sort_order"
                    defaultValue={editingPackage?.sort_order || 0}
                    className="mt-1 w-full rounded-lg border border-zinc-300 px-4 py-2 focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                  />
                </div>
                <div className="flex items-end">
                  <div className="text-sm text-zinc-500">
                    {editingPackage && (
                      <span>已售: {editingPackage.sold_count || 0}</span>
                    )}
                  </div>
                </div>
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowModal(false);
                    setEditingPackage(null);
                  }}
                  className="flex-1 rounded-lg border border-zinc-300 px-4 py-2 text-zinc-700 hover:bg-zinc-50"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending || updateMutation.isPending}
                  className="flex-1 rounded-lg bg-zinc-900 px-4 py-2 text-white hover:bg-zinc-800 disabled:opacity-50"
                >
                  {createMutation.isPending || updateMutation.isPending ? '处理中...' : '确认'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

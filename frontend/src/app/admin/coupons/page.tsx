'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { couponApi } from '@/lib/api/coupon';
import type { Coupon } from '@/types/api';
import { Plus, Edit3, Power, PowerOff } from 'lucide-react';

export default function AdminCouponsPage() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [editingCoupon, setEditingCoupon] = useState<Coupon | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'coupons', page, search, statusFilter],
    queryFn: () =>
      couponApi.list({
        page,
        page_size: 20,
        search: search || undefined,
        status: statusFilter || undefined,
      }),
  });

  const createMutation = useMutation({
    mutationFn: couponApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'coupons'] });
      setEditingCoupon(null);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) => couponApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'coupons'] });
      setEditingCoupon(null);
    },
  });

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);

    const maxUses = parseInt(formData.get('max_uses') as string, 10);
    const minAmount = parseFloat(formData.get('min_amount') as string);

    const payload = {
      code: (formData.get('code') as string).trim(),
      type: formData.get('type') as 'fixed' | 'percent',
      value: parseFloat(formData.get('value') as string),
      max_uses: isNaN(maxUses) ? 0 : maxUses,
      min_amount: isNaN(minAmount) ? 0 : minAmount,
      starts_at: (formData.get('starts_at') as string) || '',
      ends_at: (formData.get('ends_at') as string) || '',
      status: formData.get('status') as string,
    };

    if (editingCoupon) {
      updateMutation.mutate({ id: editingCoupon.id, data: payload });
    } else {
      createMutation.mutate(payload);
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
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">优惠码管理</h1>
          <p className="text-sm text-zinc-500">创建与维护优惠码规则</p>
        </div>
        <button
          onClick={() => setEditingCoupon(null)}
          className="inline-flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800"
        >
          <Plus className="h-4 w-4" />
          新建优惠码
        </button>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.6fr_1fr]">
        <div className="space-y-4">
          <div className="flex flex-col gap-3 rounded-xl border border-zinc-200 bg-white p-4 shadow-sm md:flex-row md:items-center">
            <div className="flex-1">
              <input
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
                placeholder="搜索优惠码"
                className="w-full rounded-lg border border-zinc-200 bg-zinc-50 px-4 py-2 text-sm outline-none transition-all focus:border-zinc-900 focus:bg-white focus:ring-2 focus:ring-zinc-900/10"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(1);
              }}
              className="rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm"
            >
              <option value="">全部状态</option>
              <option value="active">启用</option>
              <option value="inactive">停用</option>
            </select>
          </div>

          <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm">
            <table className="w-full text-sm">
              <thead className="border-b border-zinc-100 bg-zinc-50/50 text-left text-xs uppercase tracking-wide text-zinc-500">
                <tr>
                  <th className="px-4 py-3">优惠码</th>
                  <th className="px-4 py-3">类型</th>
                  <th className="px-4 py-3">折扣</th>
                  <th className="px-4 py-3">使用</th>
                  <th className="px-4 py-3">状态</th>
                  <th className="px-4 py-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody>
                {data?.coupons.map((coupon) => (
                  <tr key={coupon.id} className="border-b border-zinc-100 last:border-none">
                    <td className="px-4 py-3 font-medium text-zinc-900">{coupon.code}</td>
                    <td className="px-4 py-3 text-zinc-600">{coupon.type === 'percent' ? '百分比' : '固定额'}</td>
                    <td className="px-4 py-3 text-zinc-600">
                      {coupon.type === 'percent' ? `${coupon.value}%` : `$${coupon.value.toFixed(2)}`}
                      {coupon.min_amount > 0 && (
                        <span className="ml-2 text-xs text-zinc-400">满 ${coupon.min_amount.toFixed(2)}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-zinc-600">
                      {coupon.used_count}/{coupon.max_uses || '∞'}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
                          coupon.status === 'active'
                            ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20'
                            : 'bg-zinc-100 text-zinc-600'
                        }`}
                      >
                        {coupon.status === 'active' ? '启用' : '停用'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right space-x-2">
                      <button
                        onClick={() => setEditingCoupon(coupon)}
                        className="inline-flex items-center gap-1 rounded-md border border-zinc-200 px-2 py-1 text-xs text-zinc-600 hover:bg-zinc-50"
                      >
                        <Edit3 className="h-3 w-3" />
                        编辑
                      </button>
                      <button
                        onClick={() =>
                          updateMutation.mutate({
                            id: coupon.id,
                            data: { status: coupon.status === 'active' ? 'inactive' : 'active' },
                          })
                        }
                        className="inline-flex items-center gap-1 rounded-md border border-zinc-200 px-2 py-1 text-xs text-zinc-600 hover:bg-zinc-50"
                      >
                        {coupon.status === 'active' ? <PowerOff className="h-3 w-3" /> : <Power className="h-3 w-3" />}
                        {coupon.status === 'active' ? '停用' : '启用'}
                      </button>
                    </td>
                  </tr>
                ))}
                {!data?.coupons.length && (
                  <tr>
                    <td colSpan={6} className="px-4 py-6 text-center text-sm text-zinc-500">
                      暂无优惠码
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <div className="flex items-center justify-between rounded-xl border border-zinc-100 bg-white px-6 py-4 text-sm text-zinc-600">
            <span>
              当前第 {data?.pagination.page || 1} 页 / 共 {data?.pagination.total_pages || 1} 页
            </span>
            <div className="space-x-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="rounded-md border border-zinc-200 px-3 py-1.5 text-sm text-zinc-600 disabled:cursor-not-allowed disabled:opacity-40"
              >
                上一页
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page >= (data?.pagination.total_pages || 1)}
                className="rounded-md border border-zinc-200 px-3 py-1.5 text-sm text-zinc-600 disabled:cursor-not-allowed disabled:opacity-40"
              >
                下一页
              </button>
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
          <h2 className="text-lg font-semibold text-zinc-900 mb-4">
            {editingCoupon ? '编辑优惠码' : '创建优惠码'}
          </h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">优惠码</label>
              <input
                name="code"
                defaultValue={editingCoupon?.code || ''}
                placeholder="SAVE10"
                className="w-full rounded-lg border border-zinc-300 px-4 py-2 text-sm focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                required
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-700">类型</label>
                <select
                  name="type"
                  defaultValue={editingCoupon?.type || 'fixed'}
                  className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm"
                >
                  <option value="fixed">固定额</option>
                  <option value="percent">百分比</option>
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-700">折扣值</label>
                <input
                  name="value"
                  type="number"
                  step="0.01"
                  defaultValue={editingCoupon?.value ?? ''}
                  className="w-full rounded-lg border border-zinc-300 px-4 py-2 text-sm focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                  required
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-700">最大使用次数</label>
                <input
                  name="max_uses"
                  type="number"
                  defaultValue={editingCoupon?.max_uses ?? 0}
                  className="w-full rounded-lg border border-zinc-300 px-4 py-2 text-sm focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-700">最低订单金额</label>
                <input
                  name="min_amount"
                  type="number"
                  step="0.01"
                  defaultValue={editingCoupon?.min_amount ?? 0}
                  className="w-full rounded-lg border border-zinc-300 px-4 py-2 text-sm focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-700">开始时间</label>
                <input
                  name="starts_at"
                  type="text"
                  defaultValue={editingCoupon?.starts_at || ''}
                  placeholder="2025-01-01 或 2025-01-01T09:00"
                  className="w-full rounded-lg border border-zinc-300 px-4 py-2 text-sm focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-700">结束时间</label>
                <input
                  name="ends_at"
                  type="text"
                  defaultValue={editingCoupon?.ends_at || ''}
                  placeholder="2025-02-01 或 2025-02-01T09:00"
                  className="w-full rounded-lg border border-zinc-300 px-4 py-2 text-sm focus:border-zinc-900 focus:outline-none focus:ring-2 focus:ring-zinc-900/10"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">状态</label>
              <select
                name="status"
                defaultValue={editingCoupon?.status || 'active'}
                className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm"
              >
                <option value="active">启用</option>
                <option value="inactive">停用</option>
              </select>
            </div>

            <div className="flex gap-2 pt-2">
              <button
                type="submit"
                className="flex-1 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-800"
              >
                {editingCoupon ? '保存修改' : '创建优惠码'}
              </button>
              {editingCoupon && (
                <button
                  type="button"
                  onClick={() => setEditingCoupon(null)}
                  className="flex-1 rounded-lg border border-zinc-300 px-4 py-2 text-sm text-zinc-700 hover:bg-zinc-50"
                >
                  取消
                </button>
              )}
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import Link from 'next/link';
import { Search, Eye, Ban, CheckCircle, ChevronLeft, ChevronRight } from 'lucide-react';

export default function AdminUsersPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('');
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'users', page, search, status],
    queryFn: () =>
      adminApi.getUsers({
        page,
        page_size: 20,
        search: search || undefined,
        status: status || undefined,
      }),
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'active' | 'suspended' | 'banned' }) =>
      adminApi.updateUserStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 animate-pulse rounded bg-zinc-200" />
        <div className="h-20 animate-pulse rounded-xl bg-zinc-100" />
        <div className="h-96 animate-pulse rounded-xl bg-zinc-100" />
      </div>
    );
  }

  return (
    <div className="max-w-7xl space-y-6">
      <div className="flex items-center justify-between border-b border-zinc-200 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">用户管理</h1>
          <p className="text-sm text-zinc-500">管理平台用户账户</p>
        </div>
        <div className="text-sm text-zinc-500">
          共 <span className="font-semibold text-zinc-900">{data?.pagination.total || 0}</span> 个用户
        </div>
      </div>

      {/* Filters */}
      <div className="rounded-xl border border-zinc-200 bg-white p-4 shadow-sm">
        <form onSubmit={handleSearch} className="flex gap-3">
          <div className="flex-1">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-400" />
              <input
                type="text"
                placeholder="搜索用户名或邮箱..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full rounded-lg border border-zinc-200 bg-zinc-50 py-2 pl-10 pr-4 text-sm outline-none transition-all focus:border-zinc-900 focus:bg-white focus:ring-2 focus:ring-zinc-900/10"
              />
            </div>
          </div>
          <select
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
            className="rounded-lg border border-zinc-200 bg-zinc-50 px-4 py-2 text-sm outline-none transition-all focus:border-zinc-900 focus:bg-white focus:ring-2 focus:ring-zinc-900/10"
          >
            <option value="">所有状态</option>
            <option value="active">活跃</option>
            <option value="suspended">暂停</option>
            <option value="banned">封禁</option>
          </select>
          <button
            type="submit"
            className="rounded-lg bg-zinc-900 px-6 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800"
          >
            搜索
          </button>
        </form>
      </div>

      {/* Users Table */}
      <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="border-b border-zinc-100 bg-zinc-50/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
                  用户
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
                  余额
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
                  状态
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
                  角色
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-zinc-500">
                  注册时间
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-zinc-500">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-100">
              {data?.users.map((user) => (
                <tr key={user.id} className="transition-colors hover:bg-zinc-50/50">
                  <td className="whitespace-nowrap px-6 py-4">
                    <div className="flex items-center gap-3">
                      {user.avatar_url ? (
                        <img
                          src={user.avatar_url}
                          alt={user.username || user.email}
                          className="h-8 w-8 rounded-full"
                        />
                      ) : (
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-zinc-100 text-xs font-medium text-zinc-600">
                          {(user.username || user.email)[0].toUpperCase()}
                        </div>
                      )}
                      <div>
                        <div className="text-sm font-medium text-zinc-900">
                          {user.username || user.email}
                        </div>
                        <div className="text-xs text-zinc-400">{user.id.slice(0, 8)}...</div>
                      </div>
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <div className="text-sm font-medium text-zinc-900">
                      ${user.balance.toFixed(2)}
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <span
                      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        user.status === 'active'
                          ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20'
                          : user.status === 'suspended'
                          ? 'bg-amber-50 text-amber-700 ring-1 ring-amber-600/20'
                          : 'bg-red-50 text-red-700 ring-1 ring-red-600/20'
                      }`}
                    >
                      {user.status === 'active'
                        ? '活跃'
                        : user.status === 'suspended'
                        ? '暂停'
                        : '封禁'}
                    </span>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-zinc-600">
                    {user.role === 'super_admin'
                      ? '超级管理员'
                      : user.role === 'admin'
                      ? '管理员'
                      : '用户'}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-zinc-500">
                    {new Date(user.created_at).toLocaleDateString('zh-CN')}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <Link
                        href={`/admin/users/${user.id}`}
                        className="rounded-md p-1.5 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-900"
                        title="查看详情"
                      >
                        <Eye className="h-4 w-4" />
                      </Link>
                      {user.status === 'active' && user.role === 'user' && (
                        <button
                          onClick={() =>
                            updateStatusMutation.mutate({
                              id: user.id,
                              status: 'suspended',
                            })
                          }
                          className="rounded-md p-1.5 text-zinc-400 transition-colors hover:bg-amber-50 hover:text-amber-600"
                          title="暂停用户"
                        >
                          <Ban className="h-4 w-4" />
                        </button>
                      )}
                      {user.status === 'suspended' && (
                        <button
                          onClick={() =>
                            updateStatusMutation.mutate({
                              id: user.id,
                              status: 'active',
                            })
                          }
                          className="rounded-md p-1.5 text-zinc-400 transition-colors hover:bg-emerald-50 hover:text-emerald-600"
                          title="激活用户"
                        >
                          <CheckCircle className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {data && data.pagination.total_pages > 1 && (
          <div className="flex items-center justify-between border-t border-zinc-100 bg-zinc-50/30 px-6 py-4">
            <div className="text-sm text-zinc-600">
              第 <span className="font-medium text-zinc-900">{page}</span> / {data.pagination.total_pages} 页
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="inline-flex items-center gap-1 rounded-lg border border-zinc-200 bg-white px-3 py-1.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                <ChevronLeft className="h-4 w-4" />
                上一页
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page >= data.pagination.total_pages}
                className="inline-flex items-center gap-1 rounded-lg border border-zinc-200 bg-white px-3 py-1.5 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50 disabled:cursor-not-allowed disabled:opacity-40"
              >
                下一页
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

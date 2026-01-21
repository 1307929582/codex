'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, DollarSign } from 'lucide-react';
import Link from 'next/link';

export default function AdminUserDetailPage() {
  const params = useParams();
  const router = useRouter();
  const userId = params.id as string;
  const queryClient = useQueryClient();

  const [showBalanceModal, setShowBalanceModal] = useState(false);
  const [balanceAmount, setBalanceAmount] = useState('');
  const [balanceDescription, setBalanceDescription] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'user', userId],
    queryFn: () => adminApi.getUser(userId),
  });

  const updateBalanceMutation = useMutation({
    mutationFn: () =>
      adminApi.updateBalance(userId, parseFloat(balanceAmount), balanceDescription),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'user', userId] });
      setShowBalanceModal(false);
      setBalanceAmount('');
      setBalanceDescription('');
    },
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="text-center">
        <p className="text-gray-500">用户不存在</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8 flex items-center gap-4">
        <Link
          href="/admin/users"
          className="rounded-lg p-2 hover:bg-gray-100"
        >
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">用户详情</h1>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* User Info */}
        <div className="lg:col-span-2">
          <div className="rounded-lg bg-white p-6 shadow-sm">
            <h2 className="mb-4 text-xl font-bold text-gray-900">基本信息</h2>
            <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <dt className="text-sm font-medium text-gray-500">用户ID</dt>
                <dd className="mt-1 text-sm text-gray-900 font-mono">{data.user.id}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">用户名</dt>
                <dd className="mt-1 text-sm text-gray-900">{data.user.username || '未设置'}</dd>
              </div>
              {data.user.oauth_provider && (
                <div>
                  <dt className="text-sm font-medium text-gray-500">登录方式</dt>
                  <dd className="mt-1">
                    <span className="inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800">
                      {data.user.oauth_provider === 'linuxdo' ? 'LinuxDo' : data.user.oauth_provider}
                    </span>
                  </dd>
                </div>
              )}
              <div>
                <dt className="text-sm font-medium text-gray-500">LinuxDo ID</dt>
                <dd className="mt-1 text-sm text-gray-900 font-mono">
                  {data.user.oauth_provider === 'linuxdo' ? data.user.oauth_id || '未绑定' : '未绑定'}
                </dd>
              </div>
              {data.user.avatar_url && (
                <div className="sm:col-span-2">
                  <dt className="text-sm font-medium text-gray-500">头像</dt>
                  <dd className="mt-2">
                    <img
                      src={data.user.avatar_url}
                      alt="用户头像"
                      className="h-16 w-16 rounded-full border-2 border-gray-200"
                    />
                  </dd>
                </div>
              )}
              <div>
                <dt className="text-sm font-medium text-gray-500">状态</dt>
                <dd className="mt-1">
                  <span
                    className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ${
                      data.user.status === 'active'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                    }`}
                  >
                    {data.user.status === 'active' ? '活跃' : '暂停'}
                  </span>
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">角色</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {data.user.role === 'admin' ? '管理员' : '用户'}
                </dd>
              </div>
              <div className="sm:col-span-2">
                <dt className="text-sm font-medium text-gray-500">注册时间</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {new Date(data.user.created_at).toLocaleString('zh-CN')}
                </dd>
              </div>
            </dl>
          </div>
        </div>

        {/* Stats */}
        <div className="space-y-6">
          <div className="rounded-lg bg-white p-6 shadow-sm">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-500">账户余额</p>
                <p className="mt-2 text-3xl font-bold text-gray-900">
                  ${data.user.balance.toFixed(2)}
                </p>
              </div>
              <button
                onClick={() => setShowBalanceModal(true)}
                className="rounded-lg bg-blue-600 p-2 text-white hover:bg-blue-700"
              >
                <DollarSign className="h-5 w-5" />
              </button>
            </div>
          </div>

          <div className="rounded-lg bg-white p-6 shadow-sm">
            <p className="text-sm font-medium text-gray-500">API密钥数量</p>
            <p className="mt-2 text-3xl font-bold text-gray-900">
              {data.api_key_count}
            </p>
          </div>

          <div className="rounded-lg bg-white p-6 shadow-sm">
            <p className="text-sm font-medium text-gray-500">总消费</p>
            <p className="mt-2 text-3xl font-bold text-gray-900">
              ${data.total_cost.toFixed(2)}
            </p>
          </div>

          <div className="rounded-lg bg-white p-6 shadow-sm">
            <p className="text-sm font-medium text-gray-500">总Token数</p>
            <p className="mt-2 text-3xl font-bold text-gray-900">
              {data.total_tokens.toLocaleString()}
            </p>
          </div>
        </div>
      </div>

      {/* Balance Modal */}
      {showBalanceModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
          <div className="w-full max-w-md rounded-lg bg-white p-6">
            <h3 className="mb-4 text-xl font-bold text-gray-900">调整余额</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  金额（正数为充值，负数为扣除）
                </label>
                <input
                  type="number"
                  step="0.01"
                  value={balanceAmount}
                  onChange={(e) => setBalanceAmount(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  placeholder="例如: 100 或 -50"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  说明
                </label>
                <textarea
                  value={balanceDescription}
                  onChange={(e) => setBalanceDescription(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  rows={3}
                  placeholder="调整原因..."
                />
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => setShowBalanceModal(false)}
                  className="flex-1 rounded-lg border border-gray-300 px-4 py-2 text-gray-700 hover:bg-gray-50"
                >
                  取消
                </button>
                <button
                  onClick={() => updateBalanceMutation.mutate()}
                  disabled={!balanceAmount || updateBalanceMutation.isPending}
                  className="flex-1 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  {updateBalanceMutation.isPending ? '处理中...' : '确认'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

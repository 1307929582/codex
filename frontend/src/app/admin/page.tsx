'use client';

import { useQuery } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Users, DollarSign, Key, TrendingUp } from 'lucide-react';

export default function AdminDashboard() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['admin', 'overview'],
    queryFn: () => adminApi.getOverview(),
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  const cards = [
    {
      title: '总用户数',
      value: stats?.total_users || 0,
      subtitle: `活跃用户: ${stats?.active_users || 0}`,
      icon: Users,
      color: 'bg-blue-500',
    },
    {
      title: '总收入',
      value: `$${(stats?.total_revenue || 0).toFixed(2)}`,
      subtitle: `今日: $${(stats?.today_revenue || 0).toFixed(2)}`,
      icon: DollarSign,
      color: 'bg-green-500',
    },
    {
      title: '总消费',
      value: `$${(stats?.total_cost || 0).toFixed(2)}`,
      subtitle: `利润: $${((stats?.total_revenue || 0) - (stats?.total_cost || 0)).toFixed(2)}`,
      icon: TrendingUp,
      color: 'bg-purple-500',
    },
    {
      title: 'API密钥数',
      value: stats?.total_api_keys || 0,
      subtitle: `今日请求: ${stats?.today_requests || 0}`,
      icon: Key,
      color: 'bg-orange-500',
    },
  ];

  return (
    <div>
      <h1 className="mb-8 text-3xl font-bold text-gray-900">系统概览</h1>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        {cards.map((card) => (
          <div
            key={card.title}
            className="rounded-lg bg-white p-6 shadow-sm"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">
                  {card.title}
                </p>
                <p className="mt-2 text-3xl font-bold text-gray-900">
                  {card.value}
                </p>
                <p className="mt-1 text-sm text-gray-500">{card.subtitle}</p>
              </div>
              <div className={`rounded-full p-3 ${card.color}`}>
                <card.icon className="h-6 w-6 text-white" />
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-8 rounded-lg bg-white p-6 shadow-sm">
        <h2 className="mb-4 text-xl font-bold text-gray-900">快速操作</h2>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <a
            href="/admin/users"
            className="rounded-lg border border-gray-200 p-4 transition-colors hover:bg-gray-50"
          >
            <h3 className="font-medium text-gray-900">用户管理</h3>
            <p className="mt-1 text-sm text-gray-500">
              查看和管理所有用户
            </p>
          </a>
          <a
            href="/admin/settings"
            className="rounded-lg border border-gray-200 p-4 transition-colors hover:bg-gray-50"
          >
            <h3 className="font-medium text-gray-900">系统设置</h3>
            <p className="mt-1 text-sm text-gray-500">
              配置系统参数和公告
            </p>
          </a>
          <a
            href="/admin/logs"
            className="rounded-lg border border-gray-200 p-4 transition-colors hover:bg-gray-50"
          >
            <h3 className="font-medium text-gray-900">操作日志</h3>
            <p className="mt-1 text-sm text-gray-500">
              查看管理员操作记录
            </p>
          </a>
        </div>
      </div>
    </div>
  );
}

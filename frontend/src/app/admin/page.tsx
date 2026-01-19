'use client';

import { useQuery } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Users, DollarSign, Key, TrendingUp, Activity, ArrowUpRight, ArrowDownRight } from 'lucide-react';

export default function AdminDashboard() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['admin', 'overview'],
    queryFn: () => adminApi.getOverview(),
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  const cards = [
    {
      title: '总用户数',
      value: stats?.total_users || 0,
      subtitle: `活跃用户: ${stats?.active_users || 0}`,
      icon: Users,
      gradient: 'from-blue-500 to-blue-600',
      bgGradient: 'from-blue-50 to-blue-100',
      trend: '+12%',
      trendUp: true,
    },
    {
      title: '总收入',
      value: `$${(stats?.total_revenue || 0).toFixed(2)}`,
      subtitle: `今日: $${(stats?.today_revenue || 0).toFixed(2)}`,
      icon: DollarSign,
      gradient: 'from-green-500 to-emerald-600',
      bgGradient: 'from-green-50 to-emerald-100',
      trend: '+8%',
      trendUp: true,
    },
    {
      title: '总消费',
      value: `$${(stats?.total_cost || 0).toFixed(2)}`,
      subtitle: `利润: $${((stats?.total_revenue || 0) - (stats?.total_cost || 0)).toFixed(2)}`,
      icon: TrendingUp,
      gradient: 'from-purple-500 to-purple-600',
      bgGradient: 'from-purple-50 to-purple-100',
      trend: '+15%',
      trendUp: true,
    },
    {
      title: 'API密钥数',
      value: stats?.total_api_keys || 0,
      subtitle: `今日请求: ${stats?.today_requests || 0}`,
      icon: Key,
      gradient: 'from-orange-500 to-orange-600',
      bgGradient: 'from-orange-50 to-orange-100',
      trend: '+5%',
      trendUp: true,
    },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900">系统概览</h1>
        <p className="mt-2 text-sm text-gray-600">实时监控系统运行状态和关键指标</p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        {cards.map((card) => (
          <div
            key={card.title}
            className="group relative overflow-hidden rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-900/5 transition-all hover:shadow-lg hover:ring-gray-900/10"
          >
            {/* Background Gradient */}
            <div className={`absolute inset-0 bg-gradient-to-br ${card.bgGradient} opacity-0 transition-opacity group-hover:opacity-100`}></div>

            {/* Content */}
            <div className="relative">
              <div className="flex items-center justify-between">
                <div className={`rounded-xl bg-gradient-to-br ${card.gradient} p-3 shadow-lg`}>
                  <card.icon className="h-6 w-6 text-white" />
                </div>
                <div className={`flex items-center gap-1 text-sm font-medium ${card.trendUp ? 'text-green-600' : 'text-red-600'}`}>
                  {card.trendUp ? <ArrowUpRight className="h-4 w-4" /> : <ArrowDownRight className="h-4 w-4" />}
                  {card.trend}
                </div>
              </div>

              <div className="mt-4">
                <p className="text-sm font-medium text-gray-600">{card.title}</p>
                <p className="mt-2 text-3xl font-bold text-gray-900">{card.value}</p>
                <p className="mt-1 text-sm text-gray-500">{card.subtitle}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Quick Actions */}
      <div className="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-900/5">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-xl font-bold text-gray-900">快速操作</h2>
            <p className="mt-1 text-sm text-gray-600">常用管理功能快捷入口</p>
          </div>
          <Activity className="h-5 w-5 text-gray-400" />
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <a
            href="/admin/users"
            className="group relative overflow-hidden rounded-xl border border-gray-200 p-6 transition-all hover:border-blue-500 hover:shadow-md"
          >
            <div className="absolute inset-0 bg-gradient-to-br from-blue-50 to-transparent opacity-0 transition-opacity group-hover:opacity-100"></div>
            <div className="relative">
              <div className="flex items-center gap-3">
                <div className="rounded-lg bg-blue-100 p-2 group-hover:bg-blue-200 transition-colors">
                  <Users className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">用户管理</h3>
                  <p className="mt-1 text-sm text-gray-500">查看和管理所有用户</p>
                </div>
              </div>
            </div>
          </a>

          <a
            href="/admin/settings"
            className="group relative overflow-hidden rounded-xl border border-gray-200 p-6 transition-all hover:border-purple-500 hover:shadow-md"
          >
            <div className="absolute inset-0 bg-gradient-to-br from-purple-50 to-transparent opacity-0 transition-opacity group-hover:opacity-100"></div>
            <div className="relative">
              <div className="flex items-center gap-3">
                <div className="rounded-lg bg-purple-100 p-2 group-hover:bg-purple-200 transition-colors">
                  <Key className="h-5 w-5 text-purple-600" />
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">系统设置</h3>
                  <p className="mt-1 text-sm text-gray-500">配置系统参数和公告</p>
                </div>
              </div>
            </div>
          </a>

          <a
            href="/admin/logs"
            className="group relative overflow-hidden rounded-xl border border-gray-200 p-6 transition-all hover:border-orange-500 hover:shadow-md"
          >
            <div className="absolute inset-0 bg-gradient-to-br from-orange-50 to-transparent opacity-0 transition-opacity group-hover:opacity-100"></div>
            <div className="relative">
              <div className="flex items-center gap-3">
                <div className="rounded-lg bg-orange-100 p-2 group-hover:bg-orange-200 transition-colors">
                  <Activity className="h-5 w-5 text-orange-600" />
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">操作日志</h3>
                  <p className="mt-1 text-sm text-gray-500">查看管理员操作记录</p>
                </div>
              </div>
            </div>
          </a>
        </div>
      </div>
    </div>
  );
}

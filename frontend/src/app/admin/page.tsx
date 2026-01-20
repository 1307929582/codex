'use client';

import { useQuery } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Users, DollarSign, Activity, Key, ChevronRight } from 'lucide-react';

export default function AdminDashboard() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['admin', 'overview'],
    queryFn: () => adminApi.getOverview(),
  });

  const { data: chartData } = useQuery({
    queryKey: ['admin', 'usage-chart'],
    queryFn: () => adminApi.getUsageChart(),
  });

  if (isLoading) return <DashboardSkeleton />;

  const metrics = [
    {
      label: '总Token数',
      value: (stats?.total_tokens || 0).toLocaleString(),
      icon: Activity,
      desc: '累计使用Token'
    },
    {
      label: '活跃用户',
      value: stats?.active_users || 0,
      icon: Users,
      desc: `共 ${stats?.total_users || 0} 个用户`
    },
    {
      label: 'API密钥',
      value: stats?.total_api_keys || 0,
      icon: Key,
      desc: '活跃密钥'
    },
    {
      label: '今日请求',
      value: stats?.today_requests || 0,
      icon: DollarSign,
      desc: '每日流量'
    },
  ];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-zinc-900">控制台</h1>
        <p className="text-sm text-zinc-500">实时监控系统运行状态和关键指标</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric) => (
          <div
            key={metric.label}
            className="group relative overflow-hidden rounded-xl bg-white p-6 shadow-[0_2px_10px_-3px_rgba(6,81,237,0.1)] ring-1 ring-zinc-200 transition-all hover:ring-zinc-300 hover:shadow-md"
          >
            <div className="flex items-center justify-between">
              <div className="rounded-lg bg-zinc-50 p-2 ring-1 ring-zinc-100">
                <metric.icon className="h-4 w-4 text-zinc-500" />
              </div>
            </div>
            <div className="mt-4">
              <h3 className="text-sm font-medium text-zinc-500">{metric.label}</h3>
              <p className="mt-2 text-3xl font-bold tracking-tight text-zinc-900">{metric.value}</p>
              <p className="mt-1 text-xs text-zinc-400">{metric.desc}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Example Chart Area / Detailed Stats */}
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-2">
        <div className="rounded-xl bg-white p-6 shadow-sm ring-1 ring-zinc-200">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="font-semibold text-zinc-900">24小时消费趋势</h3>
            <span className="text-xs text-zinc-400">每小时</span>
          </div>
          <div className="space-y-2">
            {chartData && chartData.length > 0 ? (
              chartData.map((item) => {
                const maxCost = Math.max(...chartData.map(d => d.cost));
                const percentage = maxCost > 0 ? (item.cost / maxCost) * 100 : 0;
                return (
                  <div key={item.hour} className="flex items-center gap-3">
                    <span className="text-xs text-zinc-500 w-12">{item.hour}</span>
                    <div className="flex-1 h-6 bg-zinc-100 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-blue-500 transition-all"
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                    <span className="text-xs font-medium text-zinc-700 w-16 text-right">
                      ${item.cost.toFixed(4)}
                    </span>
                  </div>
                );
              })
            ) : (
              <div className="flex h-64 items-center justify-center rounded-lg border border-dashed border-zinc-200 bg-zinc-50/50">
                <p className="text-sm text-zinc-400">暂无数据</p>
              </div>
            )}
          </div>
        </div>

        <div className="rounded-xl bg-white p-6 shadow-sm ring-1 ring-zinc-200">
           <div className="mb-4">
            <h3 className="font-semibold text-zinc-900">快速访问</h3>
          </div>
          <div className="grid gap-3">
             {[
               { name: '管理用户', href: '/admin/users' },
               { name: '系统设置', href: '/admin/settings' },
               { name: '查看日志', href: '/admin/logs' }
             ].map((action) => (
               <a
                 key={action.name}
                 href={action.href}
                 className="flex w-full items-center justify-between rounded-lg border border-zinc-100 bg-zinc-50 px-4 py-3 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-100 hover:text-zinc-900"
               >
                 {action.name}
                 <ChevronRight className="h-4 w-4 text-zinc-400" />
               </a>
             ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-8 animate-pulse">
      <div className="h-8 w-48 rounded bg-zinc-200"></div>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="h-40 rounded-xl bg-zinc-200"></div>
        ))}
      </div>
    </div>
  )
}

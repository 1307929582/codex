'use client';

import { useQuery } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Users, DollarSign, Activity, Key } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

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

  // Transform chart data for recharts
  const formattedChartData = chartData?.map(item => ({
    hour: item.hour,
    cost: item.cost
  })) || [];

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

      {/* Usage Trend Chart */}
      <div className="rounded-xl bg-white p-6 shadow-sm ring-1 ring-zinc-200">
        <div className="mb-6">
          <h3 className="text-base font-semibold text-zinc-900">24小时使用趋势</h3>
          <p className="mt-1 text-xs text-zinc-500">近24小时每小时消费统计</p>
        </div>
        <div className="h-[300px]">
          {formattedChartData.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={formattedChartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis
                  dataKey="hour"
                  tick={{ fontSize: 12 }}
                  stroke="#888"
                />
                <YAxis
                  tick={{ fontSize: 12 }}
                  stroke="#888"
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'rgba(255, 255, 255, 0.95)',
                    border: '1px solid #e5e5ea',
                    borderRadius: '8px',
                    boxShadow: '0 2px 8px rgba(0, 0, 0, 0.08)',
                  }}
                  formatter={(value: number) => [`$${value.toFixed(4)}`, '消费']}
                />
                <Line
                  type="monotone"
                  dataKey="cost"
                  stroke="#007aff"
                  strokeWidth={2}
                  dot={{ fill: '#007aff', r: 4 }}
                  activeDot={{ r: 6 }}
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex h-full items-center justify-center rounded-lg border border-dashed border-zinc-200 bg-zinc-50/50">
              <p className="text-sm text-zinc-400">暂无数据</p>
            </div>
          )}
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

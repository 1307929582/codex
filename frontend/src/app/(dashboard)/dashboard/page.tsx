'use client';

import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import apiClient from '@/lib/api/client';
import { packageApi } from '@/lib/api/package';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { useAuthStore } from '@/lib/stores/auth';

export default function DashboardPage() {
  const { user } = useAuthStore();
  const [activeTab, setActiveTab] = useState('request');
  const displayName = user?.username || '未设置';
  const linuxdoId = user?.oauth_provider === 'linuxdo' ? user?.oauth_id : '';

  const { data: stats, isLoading } = useQuery({
    queryKey: ['usage-stats'],
    queryFn: async () => {
      const res = await apiClient.get('/api/usage/stats');
      return res.data;
    },
  });

  const { data: balance } = useQuery({
    queryKey: ['balance'],
    queryFn: async () => {
      const res = await apiClient.get('/api/account/balance');
      return res.data;
    },
  });

  const { data: dailyUsage } = useQuery({
    queryKey: ['daily-usage'],
    queryFn: () => packageApi.getDailyUsage(),
  });

  const { data: dailyTrend } = useQuery({
    queryKey: ['daily-trend', activeTab],
    queryFn: async () => {
      const typeMap: Record<string, string> = {
        request: 'requests',
        cost: 'cost',
        distribution: 'cost',
        ranking: 'cost',
      };
      const res = await apiClient.get('/api/usage/daily-trend', {
        params: { type: typeMap[activeTab] || 'cost' },
      });
      return res.data;
    },
  });

  // Mock 7-day data for chart (replace with real API data)
  const chartData = dailyTrend || [
    { date: '01-14', value: 0 },
    { date: '01-15', value: 0 },
    { date: '01-16', value: 0 },
    { date: '01-17', value: 0 },
    { date: '01-18', value: 0 },
    { date: '01-19', value: 0 },
    { date: '01-20', value: 0 },
  ];

  const chartTotal = chartData.reduce((sum: number, item: any) => sum + (item.value || 0), 0);
  const dailyLimit = dailyUsage?.global_daily_limit ?? null;
  const totalUsed = dailyUsage?.total_used_amount ?? 0;
  const dailyRemaining =
    dailyUsage?.global_remaining ?? (dailyLimit !== null ? Math.max(dailyLimit - totalUsed, 0) : 0);

  if (isLoading) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <div className="text-muted-foreground">加载中...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header Card */}
      <Card className="border-none shadow-none">
        <CardHeader className="pb-3">
          <CardTitle className="text-2xl font-semibold">仪表盘</CardTitle>
          <p className="text-sm text-muted-foreground">
            欢迎回来，{displayName}
            {linuxdoId && <span className="ml-2 text-xs text-zinc-500">LinuxDo ID: {linuxdoId}</span>}
          </p>
        </CardHeader>
      </Card>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {/* Account Balance */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              账户余额
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              ${balance?.balance?.toFixed(2) || '0.00'}
            </div>
          </CardContent>
        </Card>

        {/* 7-Day Requests */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              7日请求
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.seven_days_requests || 0}</div>
          </CardContent>
        </Card>

        {/* 7-Day Cost */}
        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                7日消费
              </CardTitle>
              <Button variant="ghost" size="sm" className="h-6 px-2 text-xs">
                全部
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">${stats?.seven_days_cost?.toFixed(2) || '0.00'}</div>
          </CardContent>
        </Card>

        {/* 7-Day Tokens */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              7日Token
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.seven_days_tokens || 0}</div>
          </CardContent>
        </Card>
      </div>

      {/* Usage Analysis Chart */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-base font-semibold">
                使用数据分析
              </CardTitle>
              <p className="mt-1 text-xs text-muted-foreground">近7天</p>
            </div>
            <div className="flex gap-2">
              <Button
                variant={activeTab === 'request' ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setActiveTab('request')}
                className="h-8 text-xs"
              >
                请求趋势
              </Button>
              <Button
                variant={activeTab === 'cost' ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setActiveTab('cost')}
                className="h-8 text-xs"
              >
                消费趋势
              </Button>
              <Button
                variant={activeTab === 'distribution' ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setActiveTab('distribution')}
                className="h-8 text-xs"
              >
                模型分布
              </Button>
              <Button
                variant={activeTab === 'ranking' ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setActiveTab('ranking')}
                className="h-8 text-xs"
              >
                消费排行
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="mb-4">
            <div className="text-lg font-semibold">
              {activeTab === 'request' && '请求趋势'}
              {activeTab === 'cost' && '消费趋势'}
              {activeTab === 'distribution' && '模型分布'}
              {activeTab === 'ranking' && '消费排行'}
            </div>
          </div>
          <div className="h-[300px]">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis
                  dataKey="date"
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
                />
                <Bar
                  dataKey="value"
                  fill="#007aff"
                  radius={[6, 6, 0, 0]}
                />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 text-right text-sm text-muted-foreground">
            总计: ${activeTab === 'request' ? chartTotal.toFixed(0) : chartTotal.toFixed(4)}
          </div>
        </CardContent>
      </Card>

      {/* Package Status */}
      {dailyLimit !== null && (
        <Card className="border-apple-blue/20 bg-white">
          <CardHeader>
            <CardTitle className="text-base font-semibold text-apple-blue">
              全局每日上限
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-muted-foreground">今日已用</span>
                <span className="font-medium">
                  ${totalUsed.toFixed(4)}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-muted-foreground">每日上限</span>
                <span className="font-medium">
                  ${dailyLimit.toFixed(2)}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-muted-foreground">剩余</span>
                <span className="font-medium">
                  ${dailyRemaining.toFixed(4)}
                </span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-apple-blue/10">
                <div
                  className="h-full bg-apple-blue transition-all"
                  style={{
                    width: `${Math.min(
                      ((totalUsed || 0) / (dailyLimit || 1)) * 100,
                      100
                    )}%`,
                  }}
                />
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {dailyUsage?.package && (
        <Card className="border-apple-blue/20 bg-apple-blue/5">
          <CardHeader>
            <CardTitle className="text-base font-semibold text-apple-blue">
              当前套餐
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">
                  {dailyUsage.package.package_name}
                </span>
                <span className="rounded-full bg-apple-blue/10 px-2.5 py-0.5 text-xs font-medium text-apple-blue">
                  活跃中
                </span>
              </div>

              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">今日已用</span>
                  <span className="font-medium">
                    ${dailyUsage.used_amount?.toFixed(4) || '0.00'}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">每日限额</span>
                  <span className="font-medium">
                    ${dailyUsage.daily_limit?.toFixed(2) || '0.00'}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">剩余额度</span>
                  <span className="font-medium">
                    ${dailyUsage.remaining?.toFixed(4) || '0.00'}
                  </span>
                </div>
              </div>

              <div className="h-2 w-full overflow-hidden rounded-full bg-apple-blue/10">
                <div
                  className="h-full bg-apple-blue transition-all"
                  style={{
                    width: `${Math.min(
                      ((dailyUsage.used_amount || 0) / (dailyUsage.daily_limit || 1)) * 100,
                      100
                    )}%`,
                  }}
                />
              </div>

              <div className="text-xs text-muted-foreground">
                有效期至 {new Date(dailyUsage.package.end_date).toLocaleDateString('zh-CN')}
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* No Package Promotion */}
      {!dailyUsage?.package && (
        <Card className="border-apple-blue/20 bg-apple-blue/5">
          <CardHeader>
            <CardTitle className="text-base font-semibold text-apple-blue">
              订阅套餐
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="mb-4 text-sm text-muted-foreground">
              购买套餐享受每日固定额度，超出部分自动从余额扣费，更加灵活便捷。
            </p>
            <Button className="w-full" asChild>
              <a href="/packages">查看套餐</a>
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

'use client';

import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/api/client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { DollarSign, Activity, TrendingUp, Key } from 'lucide-react';

export default function DashboardPage() {
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

  const { data: keys } = useQuery({
    queryKey: ['keys'],
    queryFn: async () => {
      const res = await apiClient.get('/api/keys');
      return res.data;
    },
  });

  if (isLoading) {
    return <div>加载中...</div>;
  }

  const activeKeysCount = keys?.filter((k: any) => k.status === 'active').length || 0;

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">控制台</h2>
        <p className="text-muted-foreground">欢迎使用 Codex Gateway</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">账户余额</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${balance?.balance?.toFixed(2) || '0.00'}</div>
            <p className="text-xs text-muted-foreground">可用额度</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日消费</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${stats?.today_cost?.toFixed(4) || '0.00'}</div>
            <p className="text-xs text-muted-foreground">今日使用量</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">本月消费</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${stats?.month_cost?.toFixed(2) || '0.00'}</div>
            <p className="text-xs text-muted-foreground">本月累计</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">活跃密钥</CardTitle>
            <Key className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeKeysCount}</div>
            <p className="text-xs text-muted-foreground">API密钥数量</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>统计概览</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">总消费（全部时间）</span>
              <span className="text-sm font-bold">${stats?.total_cost?.toFixed(2) || '0.00'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">API密钥总数</span>
              <span className="text-sm font-bold">{keys?.length || 0}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

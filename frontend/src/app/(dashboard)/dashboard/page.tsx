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
    return <div>Loading...</div>;
  }

  const activeKeysCount = keys?.filter((k: any) => k.status === 'active').length || 0;

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
        <p className="text-muted-foreground">Welcome to your Codex Gateway dashboard</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Balance</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${balance?.balance?.toFixed(2) || '0.00'}</div>
            <p className="text-xs text-muted-foreground">Available credit</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Today's Cost</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${stats?.today_cost?.toFixed(4) || '0.00'}</div>
            <p className="text-xs text-muted-foreground">Usage today</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Month's Cost</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${stats?.month_cost?.toFixed(2) || '0.00'}</div>
            <p className="text-xs text-muted-foreground">This month</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Keys</CardTitle>
            <Key className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeKeysCount}</div>
            <p className="text-xs text-muted-foreground">API keys</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Quick Stats</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">Total Cost (All Time)</span>
              <span className="text-sm font-bold">${stats?.total_cost?.toFixed(2) || '0.00'}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">Total API Keys</span>
              <span className="text-sm font-bold">{keys?.length || 0}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

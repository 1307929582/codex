'use client';

import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/api/client';
import { packageApi } from '@/lib/api/package';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { DollarSign, Activity, TrendingUp, Key, Package, Calendar } from 'lucide-react';
import Link from 'next/link';

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

  const { data: dailyUsage } = useQuery({
    queryKey: ['daily-usage'],
    queryFn: () => packageApi.getDailyUsage(),
  });

  if (isLoading) {
    return <div>加载中...</div>;
  }

  const activeKeysCount = keys?.filter((k: any) => k.status === 'active').length || 0;

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">控制台</h2>
        <p className="text-muted-foreground">欢迎使用 Zenscale Codex</p>
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

      {/* Package Status */}
      {dailyUsage?.package && (
        <Card className="border-emerald-200 bg-emerald-50/50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-emerald-900">
              <Package className="h-5 w-5" />
              当前套餐
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-emerald-900">
                  {dailyUsage.package.package_name}
                </span>
                <span className="rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-medium text-emerald-700">
                  活跃中
                </span>
              </div>

              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-emerald-700">今日已用</span>
                  <span className="font-medium text-emerald-900">
                    ${dailyUsage.used_amount?.toFixed(4) || '0.00'}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-emerald-700">每日限额</span>
                  <span className="font-medium text-emerald-900">
                    ${dailyUsage.daily_limit?.toFixed(2) || '0.00'}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-emerald-700">剩余额度</span>
                  <span className="font-medium text-emerald-900">
                    ${dailyUsage.remaining?.toFixed(4) || '0.00'}
                  </span>
                </div>
              </div>

              <div className="h-2 w-full overflow-hidden rounded-full bg-emerald-100">
                <div
                  className="h-full bg-emerald-500 transition-all"
                  style={{
                    width: `${Math.min(
                      ((dailyUsage.used_amount || 0) / (dailyUsage.daily_limit || 1)) * 100,
                      100
                    )}%`,
                  }}
                />
              </div>

              <div className="flex items-center gap-2 text-xs text-emerald-700">
                <Calendar className="h-3 w-3" />
                <span>
                  有效期至 {new Date(dailyUsage.package.end_date).toLocaleDateString('zh-CN')}
                </span>
              </div>

              <Link
                href="/packages"
                className="block w-full rounded-lg border border-emerald-300 bg-white px-4 py-2 text-center text-sm font-medium text-emerald-700 transition-colors hover:bg-emerald-50"
              >
                查看更多套餐
              </Link>
            </div>
          </CardContent>
        </Card>
      )}

      {/* No Package - Promote Packages */}
      {!dailyUsage?.package && (
        <Card className="border-blue-200 bg-blue-50/50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-blue-900">
              <Package className="h-5 w-5" />
              购买套餐
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="mb-4 text-sm text-blue-700">
              购买套餐享受每日固定额度，超出部分自动从余额扣费，更加灵活便捷。
            </p>
            <Link
              href="/packages"
              className="block w-full rounded-lg bg-blue-600 px-4 py-2 text-center text-sm font-medium text-white transition-colors hover:bg-blue-700"
            >
              查看套餐
            </Link>
          </CardContent>
        </Card>
      )}

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

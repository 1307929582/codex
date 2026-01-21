'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { ChevronLeft, ChevronRight } from 'lucide-react';

export default function AdminUsagePage() {
  const [page, setPage] = useState(1);
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [userQuery, setUserQuery] = useState('');
  const [model, setModel] = useState('');
  const [statusCode, setStatusCode] = useState('');
  const [apiKeyId, setApiKeyId] = useState('');
  const pageSize = 20;

  useEffect(() => {
    setPage(1);
  }, [startDate, endDate, userQuery, model, statusCode, apiKeyId]);

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'usage-logs', page, startDate, endDate, userQuery, model, statusCode, apiKeyId],
    queryFn: async () => {
      const statusCodeValue = statusCode ? Number(statusCode) : undefined;
      const apiKeyIdValue = apiKeyId ? Number(apiKeyId) : undefined;
      return adminApi.getUsageLogs({
        page,
        page_size: pageSize,
        start_date: startDate || undefined,
        end_date: endDate || undefined,
        user: userQuery || undefined,
        model: model || undefined,
        status_code: Number.isFinite(statusCodeValue) ? statusCodeValue : undefined,
        api_key_id: Number.isFinite(apiKeyIdValue) ? apiKeyIdValue : undefined,
      });
    },
  });

  const clearFilters = () => {
    setStartDate('');
    setEndDate('');
    setUserQuery('');
    setModel('');
    setStatusCode('');
    setApiKeyId('');
  };

  return (
    <div className="max-w-6xl space-y-6">
      <div className="flex items-center justify-between border-b border-zinc-200 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">使用记录</h1>
          <p className="text-sm text-zinc-500">按时间与用户维度查看API使用情况</p>
        </div>
        <div className="text-sm text-zinc-500">
          共 <span className="font-semibold text-zinc-900">{data?.pagination.total || 0}</span> 条记录
        </div>
      </div>

      <div className="rounded-xl border border-zinc-200 bg-white p-4 shadow-sm">
        <div className="grid gap-4 md:grid-cols-3">
          <div className="space-y-1">
            <label className="text-sm text-zinc-600">开始日期</label>
            <Input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-600">结束日期</label>
            <Input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-600">用户</label>
            <Input
              placeholder="ID / 邮箱 / 用户名"
              value={userQuery}
              onChange={(e) => setUserQuery(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-600">模型</label>
            <Input
              placeholder="gpt-5.1-codex"
              value={model}
              onChange={(e) => setModel(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-600">状态码</label>
            <Input
              type="number"
              placeholder="200"
              value={statusCode}
              onChange={(e) => setStatusCode(e.target.value)}
              min={100}
              max={599}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-600">API Key ID</label>
            <Input
              type="number"
              placeholder="1"
              value={apiKeyId}
              onChange={(e) => setApiKeyId(e.target.value)}
              min={1}
            />
          </div>
        </div>
        <div className="mt-4 flex items-center justify-between text-sm text-zinc-500">
          <span>支持按日期、用户、模型、状态码、API Key 过滤</span>
          <Button variant="outline" size="sm" onClick={clearFilters}>
            清除筛选
          </Button>
        </div>
      </div>

      <div className="rounded-xl border border-zinc-200 bg-white shadow-sm">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>时间</TableHead>
              <TableHead>用户</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>API Key</TableHead>
              <TableHead>输入</TableHead>
              <TableHead>输出</TableHead>
              <TableHead>缓存</TableHead>
              <TableHead>总计</TableHead>
              <TableHead>费用</TableHead>
              <TableHead>延迟</TableHead>
              <TableHead>状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={11} className="text-center h-24">
                  加载中...
                </TableCell>
              </TableRow>
            ) : data?.logs?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={11} className="text-center h-24">
                  未找到使用记录。
                </TableCell>
              </TableRow>
            ) : (
              data?.logs?.map((log) => (
                <TableRow key={log.request_id}>
                  <TableCell className="text-xs text-zinc-600">
                    {new Date(log.created_at).toLocaleString('zh-CN')}
                  </TableCell>
                  <TableCell>
                    <div className="text-sm font-medium text-zinc-900">
                      {log.user_email || log.user_id}
                    </div>
                    {log.username && (
                      <div className="text-xs text-zinc-500">{log.username}</div>
                    )}
                    <div className="text-xs text-zinc-400 font-mono">{log.user_id}</div>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{log.model}</TableCell>
                  <TableCell className="text-sm">{log.api_key_id}</TableCell>
                  <TableCell>{log.input_tokens}</TableCell>
                  <TableCell>{log.output_tokens}</TableCell>
                  <TableCell>
                    {log.cached_tokens > 0 ? (
                      <span className="text-emerald-600 font-medium">{log.cached_tokens}</span>
                    ) : (
                      <span className="text-zinc-400">0</span>
                    )}
                  </TableCell>
                  <TableCell className="font-medium">{log.total_tokens}</TableCell>
                  <TableCell className="font-medium">${log.cost.toFixed(4)}</TableCell>
                  <TableCell>{log.latency_ms}ms</TableCell>
                  <TableCell>
                    <Badge variant={log.status_code === 200 ? 'success' : 'destructive'}>
                      {log.status_code}
                    </Badge>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {data && data.pagination.total_pages > 1 && (
        <div className="flex items-center justify-between rounded-xl border border-zinc-200 bg-white px-6 py-4 shadow-sm">
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
  );
}

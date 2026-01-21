'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/api/client';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ChevronLeft, ChevronRight } from 'lucide-react';

export default function UsagePage() {
  const [page, setPage] = useState(1);
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const pageSize = 20;

  const getBillableInputTokens = (inputTokens: number, cachedTokens: number) => {
    if (cachedTokens <= 0) {
      return inputTokens;
    }
    if (cachedTokens > inputTokens) {
      return inputTokens;
    }
    return inputTokens - cachedTokens;
  };

  useEffect(() => {
    setPage(1);
  }, [startDate, endDate]);

  const { data, isLoading } = useQuery({
    queryKey: ['usage-logs', page, startDate, endDate],
    queryFn: async () => {
      const params = new URLSearchParams({
        page: page.toString(),
        page_size: pageSize.toString(),
      });
      if (startDate) {
        params.set('start_date', startDate);
      }
      if (endDate) {
        params.set('end_date', endDate);
      }
      const res = await apiClient.get(`/api/usage/logs?${params.toString()}`);
      return res.data;
    },
  });

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">使用记录</h2>
        <p className="text-muted-foreground">查看您的API使用历史</p>
      </div>

      <div className="rounded-md border bg-white p-4">
        <div className="flex flex-wrap items-end gap-4">
          <div className="space-y-1">
            <label className="text-sm text-muted-foreground">开始日期</label>
            <Input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="w-40"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-muted-foreground">结束日期</label>
            <Input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="w-40"
            />
          </div>
          <Button
            variant="outline"
            onClick={() => {
              setStartDate('');
              setEndDate('');
            }}
          >
            清除筛选
          </Button>
        </div>
      </div>

      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>时间</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>计费输入</TableHead>
              <TableHead>输出Token</TableHead>
              <TableHead>缓存Token</TableHead>
              <TableHead>总Token</TableHead>
              <TableHead>费用</TableHead>
              <TableHead>延迟</TableHead>
              <TableHead>状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={9} className="text-center h-24">
                  加载中...
                </TableCell>
              </TableRow>
            ) : data?.data?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9} className="text-center h-24">
                  未找到使用记录。
                </TableCell>
              </TableRow>
            ) : (
              data?.data?.map((log: any) => (
                <TableRow key={log.request_id}>
                  <TableCell className="text-sm">
                    {new Date(log.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{log.model}</TableCell>
                  <TableCell>
                    {getBillableInputTokens(log.input_tokens, log.cached_tokens)}
                  </TableCell>
                  <TableCell>{log.output_tokens}</TableCell>
                  <TableCell>
                    {log.cached_tokens > 0 ? (
                      <span className="text-green-600 font-medium">{log.cached_tokens}</span>
                    ) : (
                      <span className="text-gray-400">0</span>
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

      {data && data.total_page > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            第 {data.page} 页，共 {data.total_page} 页（共 {data.total} 条记录）
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              <ChevronLeft className="h-4 w-4" />
              上一页
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.min(data.total_page, p + 1))}
              disabled={page === data.total_page}
            >
              下一页
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

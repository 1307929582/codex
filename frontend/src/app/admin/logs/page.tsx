'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Activity, ChevronLeft, ChevronRight, Clock } from 'lucide-react';

export default function AdminLogsPage() {
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ['admin', 'logs', page],
    queryFn: () => adminApi.getLogs({ page, page_size: 50 }),
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 animate-pulse rounded bg-zinc-200" />
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="h-24 animate-pulse rounded-xl bg-zinc-100" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl space-y-6">
      <div className="flex items-center justify-between border-b border-zinc-200 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">操作日志</h1>
          <p className="text-sm text-zinc-500">系统操作审计记录</p>
        </div>
        <div className="text-sm text-zinc-500">
          共 <span className="font-semibold text-zinc-900">{data?.pagination.total || 0}</span> 条记录
        </div>
      </div>

      {/* Timeline View */}
      <div className="space-y-4">
        {data?.logs.map((log: any, index: number) => (
          <div
            key={log.id}
            className="group relative rounded-xl border border-zinc-200 bg-white p-5 shadow-sm transition-all hover:shadow-md"
          >
            {/* Timeline Connector */}
            {index !== data.logs.length - 1 && (
              <div className="absolute left-[2.125rem] top-[3.5rem] h-[calc(100%+1rem)] w-px bg-zinc-200" />
            )}

            <div className="flex gap-4">
              {/* Icon */}
              <div className="relative flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-zinc-100 ring-4 ring-white">
                <Activity className="h-4 w-4 text-zinc-600" />
              </div>

              {/* Content */}
              <div className="flex-1 space-y-2">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="inline-flex rounded-md bg-zinc-900 px-2.5 py-0.5 text-xs font-medium text-white">
                        {log.action}
                      </span>
                      {log.target && (
                        <span className="text-sm text-zinc-600">
                          → <span className="font-medium text-zinc-900">{log.target}</span>
                        </span>
                      )}
                    </div>
                    {log.details && (
                      <p className="mt-2 text-sm text-zinc-600">{log.details}</p>
                    )}
                  </div>
                  <div className="flex items-center gap-1.5 text-xs text-zinc-400">
                    <Clock className="h-3.5 w-3.5" />
                    {new Date(log.created_at).toLocaleString('zh-CN', {
                      month: 'short',
                      day: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </div>
                </div>

                {/* Metadata */}
                {log.ip_address && (
                  <div className="flex items-center gap-4 text-xs text-zinc-400">
                    <span className="inline-flex items-center gap-1">
                      <span className="h-1.5 w-1.5 rounded-full bg-zinc-300" />
                      IP: {log.ip_address}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Pagination */}
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

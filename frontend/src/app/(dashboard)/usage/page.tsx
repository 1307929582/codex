'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/api/client';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ChevronLeft, ChevronRight } from 'lucide-react';

export default function UsagePage() {
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { data, isLoading } = useQuery({
    queryKey: ['usage-logs', page],
    queryFn: async () => {
      const res = await apiClient.get(`/api/usage/logs?page=${page}&page_size=${pageSize}`);
      return res.data;
    },
  });

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Usage Logs</h2>
        <p className="text-muted-foreground">View your API usage history</p>
      </div>

      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Model</TableHead>
              <TableHead>Input Tokens</TableHead>
              <TableHead>Output Tokens</TableHead>
              <TableHead>Total Tokens</TableHead>
              <TableHead>Cost</TableHead>
              <TableHead>Latency</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center h-24">
                  Loading logs...
                </TableCell>
              </TableRow>
            ) : data?.data?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center h-24">
                  No usage logs found.
                </TableCell>
              </TableRow>
            ) : (
              data?.data?.map((log: any) => (
                <TableRow key={log.request_id}>
                  <TableCell className="text-sm">
                    {new Date(log.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell className="font-mono text-xs">{log.model}</TableCell>
                  <TableCell>{log.input_tokens}</TableCell>
                  <TableCell>{log.output_tokens}</TableCell>
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
            Page {data.page} of {data.total_page} ({data.total} total records)
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
            >
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage(p => Math.min(data.total_page, p + 1))}
              disabled={page === data.total_page}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

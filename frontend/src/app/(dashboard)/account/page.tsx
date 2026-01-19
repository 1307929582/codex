'use client';

import { useQuery } from '@tanstack/react-query';
import apiClient from '@/lib/api/client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { DollarSign } from 'lucide-react';

export default function AccountPage() {
  const { data: balance, isLoading: balanceLoading } = useQuery({
    queryKey: ['balance'],
    queryFn: async () => {
      const res = await apiClient.get('/api/account/balance');
      return res.data;
    },
  });

  const { data: transactions, isLoading: transactionsLoading } = useQuery({
    queryKey: ['transactions'],
    queryFn: async () => {
      const res = await apiClient.get('/api/account/transactions');
      return res.data;
    },
  });

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">账户信息</h2>
        <p className="text-muted-foreground">管理您的账户和账单</p>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle>当前余额</CardTitle>
          <DollarSign className="h-6 w-6 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          {balanceLoading ? (
            <div>加载中...</div>
          ) : (
            <div>
              <div className="text-4xl font-bold">${balance?.balance?.toFixed(2) || '0.00'}</div>
              <p className="text-sm text-muted-foreground mt-2">
                货币: {balance?.currency || 'USD'}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      <div>
        <h3 className="text-xl font-semibold mb-4">交易历史</h3>
        <div className="rounded-md border bg-white">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>日期</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>描述</TableHead>
                <TableHead className="text-right">金额</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {transactionsLoading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center h-24">
                    加载中...
                  </TableCell>
                </TableRow>
              ) : transactions?.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center h-24">
                    未找到交易记录。
                  </TableCell>
                </TableRow>
              ) : (
                transactions?.map((tx: any) => (
                  <TableRow key={tx.id}>
                    <TableCell>{new Date(tx.created_at).toLocaleString()}</TableCell>
                    <TableCell>
                      <Badge variant={tx.type === 'deposit' ? 'success' : 'secondary'}>
                        {tx.type === 'deposit' ? '充值' : tx.type}
                      </Badge>
                    </TableCell>
                    <TableCell>{tx.description}</TableCell>
                    <TableCell className="text-right font-medium">
                      {tx.type === 'deposit' ? '+' : '-'}${Math.abs(tx.amount).toFixed(2)}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </div>
  );
}

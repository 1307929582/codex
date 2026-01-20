'use client';

import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import apiClient from '@/lib/api/client';
import { rechargeApi } from '@/lib/api/recharge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { DollarSign, Plus, X } from 'lucide-react';

export default function AccountPage() {
  const queryClient = useQueryClient();
  const [showRechargeDialog, setShowRechargeDialog] = useState(false);
  const [rechargeAmount, setRechargeAmount] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

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

  const handleRecharge = async () => {
    const amount = parseFloat(rechargeAmount);

    if (isNaN(amount) || amount <= 0) {
      alert('请输入有效的充值金额');
      return;
    }

    try {
      setIsProcessing(true);
      const result = await rechargeApi.createOrder(amount);

      // Create form and submit to Credit payment
      const form = document.createElement('form');
      form.method = 'POST';
      form.action = result.payment_url;

      Object.entries(result.params).forEach(([key, value]) => {
        const input = document.createElement('input');
        input.type = 'hidden';
        input.name = key;
        input.value = value;
        form.appendChild(input);
      });

      document.body.appendChild(form);
      form.submit();
    } catch (error: any) {
      console.error('Recharge failed:', error);
      const errorMessage = error?.response?.data?.error || error?.message || '充值失败';
      alert(`充值失败: ${errorMessage}`);
      setIsProcessing(false);
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">账户信息</h2>
        <p className="text-muted-foreground">管理您的账户和账单</p>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle>当前余额</CardTitle>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowRechargeDialog(true)}
              className="inline-flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800"
            >
              <Plus className="h-4 w-4" />
              充值
            </button>
            <DollarSign className="h-6 w-6 text-muted-foreground" />
          </div>
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

      {/* Recharge Dialog */}
      {showRechargeDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-zinc-900">充值余额</h3>
              <button
                onClick={() => setShowRechargeDialog(false)}
                className="rounded-lg p-1 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-600"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="mb-2 block text-sm font-medium text-zinc-700">
                  充值金额 (USD)
                </label>
                <input
                  type="number"
                  value={rechargeAmount}
                  onChange={(e) => setRechargeAmount(e.target.value)}
                  placeholder="请输入充值金额"
                  min="0"
                  step="0.01"
                  className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                />
                <p className="mt-1 text-xs text-zinc-500">
                  最低充值金额: $10.00
                </p>
              </div>

              {/* Quick Amount Buttons */}
              <div className="grid grid-cols-4 gap-2">
                {[10, 20, 50, 100].map((amount) => (
                  <button
                    key={amount}
                    onClick={() => setRechargeAmount(amount.toString())}
                    className="rounded-lg border border-zinc-200 px-3 py-2 text-sm font-medium text-zinc-700 transition-colors hover:border-zinc-900 hover:bg-zinc-50"
                  >
                    ${amount}
                  </button>
                ))}
              </div>

              <div className="rounded-lg bg-blue-50 p-4 text-sm text-blue-700">
                <p className="font-medium mb-1">支付说明</p>
                <ul className="text-xs space-y-1 list-disc list-inside">
                  <li>支持使用 Linux Do Credit 支付</li>
                  <li>充值成功后余额立即到账</li>
                  <li>充值金额可用于按量付费和套餐超额扣费</li>
                </ul>
              </div>

              <div className="flex gap-3">
                <button
                  onClick={() => setShowRechargeDialog(false)}
                  className="flex-1 rounded-lg border border-zinc-200 px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
                >
                  取消
                </button>
                <button
                  onClick={handleRecharge}
                  disabled={isProcessing}
                  className="flex-1 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isProcessing ? '处理中...' : '确认充值'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

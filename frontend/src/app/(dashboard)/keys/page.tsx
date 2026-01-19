'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Plus, Trash2, Copy, Check } from 'lucide-react';

export default function KeysPage() {
  const queryClient = useQueryClient();
  const [newKeyName, setNewKeyName] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [copiedId, setCopiedId] = useState<number | null>(null);

  const { data: keys, isLoading } = useQuery({
    queryKey: ['keys'],
    queryFn: async () => {
      const res = await apiClient.get('/api/keys');
      return res.data;
    },
  });

  const createKeyMutation = useMutation({
    mutationFn: async (name: string) => {
      const res = await apiClient.post('/api/keys', { name });
      return res.data;
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['keys'] });
      setNewKeyName('');
      setNewKey(data.key);
    },
  });

  const deleteKeyMutation = useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/api/keys/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['keys'] });
    },
  });

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    if (newKeyName.trim()) {
      createKeyMutation.mutate(newKeyName);
    }
  };

  const copyToClipboard = (text: string, id: number) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">API密钥</h2>
          <p className="text-muted-foreground">管理您的API密钥</p>
        </div>
        <Button onClick={() => setIsCreating(!isCreating)}>
          <Plus className="mr-2 h-4 w-4" /> 创建新密钥
        </Button>
      </div>

      {isCreating && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">创建新的API密钥</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreate} className="flex gap-4">
              <Input
                placeholder="密钥名称（例如：生产环境应用）"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                className="max-w-md"
              />
              <Button type="submit" isLoading={createKeyMutation.isPending}>
                创建
              </Button>
              <Button type="button" variant="outline" onClick={() => setIsCreating(false)}>
                取消
              </Button>
            </form>
          </CardContent>
        </Card>
      )}

      {newKey && (
        <Card className="border-green-200 bg-green-50">
          <CardHeader>
            <CardTitle className="text-lg text-green-900">API密钥创建成功！</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-green-800 mb-2">
              ⚠️ 请立即保存此密钥！它不会再次显示。
            </p>
            <div className="flex gap-2">
              <Input value={newKey} readOnly className="font-mono text-sm" />
              <Button onClick={() => copyToClipboard(newKey, -1)} variant="outline">
                <Copy className="h-4 w-4" />
              </Button>
            </div>
            <Button className="mt-4" onClick={() => setNewKey(null)}>
              我已保存密钥
            </Button>
          </CardContent>
        </Card>
      )}

      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>密钥前缀</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>使用量</TableHead>
              <TableHead>创建时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center h-24">
                  加载中...
                </TableCell>
              </TableRow>
            ) : keys?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center h-24">
                  未找到API密钥。创建一个开始使用。
                </TableCell>
              </TableRow>
            ) : (
              keys?.map((key: any) => (
                <TableRow key={key.id}>
                  <TableCell className="font-medium">{key.name}</TableCell>
                  <TableCell className="font-mono text-xs">{key.key_prefix}...</TableCell>
                  <TableCell>
                    <Badge variant={key.status === 'active' ? 'success' : 'secondary'}>
                      {key.status === 'active' ? '活跃' : key.status}
                    </Badge>
                  </TableCell>
                  <TableCell>{key.total_usage} tokens</TableCell>
                  <TableCell>{new Date(key.created_at).toLocaleDateString()}</TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => deleteKeyMutation.mutate(key.id)}
                      disabled={deleteKeyMutation.isPending}
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

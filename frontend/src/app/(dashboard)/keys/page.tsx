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
          <h2 className="text-3xl font-bold tracking-tight">API Keys</h2>
          <p className="text-muted-foreground">Manage your API keys</p>
        </div>
        <Button onClick={() => setIsCreating(!isCreating)}>
          <Plus className="mr-2 h-4 w-4" /> Create New Key
        </Button>
      </div>

      {isCreating && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Create New API Key</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreate} className="flex gap-4">
              <Input
                placeholder="Key Name (e.g. Production App)"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                className="max-w-md"
              />
              <Button type="submit" isLoading={createKeyMutation.isPending}>
                Create
              </Button>
              <Button type="button" variant="outline" onClick={() => setIsCreating(false)}>
                Cancel
              </Button>
            </form>
          </CardContent>
        </Card>
      )}

      {newKey && (
        <Card className="border-green-200 bg-green-50">
          <CardHeader>
            <CardTitle className="text-lg text-green-900">API Key Created!</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-green-800 mb-2">
              ⚠️ Save this key now! It won't be shown again.
            </p>
            <div className="flex gap-2">
              <Input value={newKey} readOnly className="font-mono text-sm" />
              <Button onClick={() => copyToClipboard(newKey, -1)} variant="outline">
                <Copy className="h-4 w-4" />
              </Button>
            </div>
            <Button className="mt-4" onClick={() => setNewKey(null)}>
              I've saved my key
            </Button>
          </CardContent>
        </Card>
      )}

      <div className="rounded-md border bg-white">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Key Prefix</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Usage</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center h-24">
                  Loading keys...
                </TableCell>
              </TableRow>
            ) : keys?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center h-24">
                  No API keys found. Create one to get started.
                </TableCell>
              </TableRow>
            ) : (
              keys?.map((key: any) => (
                <TableRow key={key.id}>
                  <TableCell className="font-medium">{key.name}</TableCell>
                  <TableCell className="font-mono text-xs">{key.key_prefix}...</TableCell>
                  <TableCell>
                    <Badge variant={key.status === 'active' ? 'success' : 'secondary'}>
                      {key.status}
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

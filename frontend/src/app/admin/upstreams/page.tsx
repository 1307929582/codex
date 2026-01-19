'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Plus, Server, Activity, Trash2, Edit, Power, PowerOff, AlertCircle, CheckCircle2, Clock } from 'lucide-react';

export default function CodexUpstreamsPage() {
  const queryClient = useQueryClient();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingUpstream, setEditingUpstream] = useState<any>(null);
  const [healthCheckMessage, setHealthCheckMessage] = useState<string | null>(null);

  const { data: upstreamsData, isLoading } = useQuery({
    queryKey: ['admin', 'codex-upstreams'],
    queryFn: () => adminApi.getCodexUpstreams(),
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  const { data: healthData } = useQuery({
    queryKey: ['admin', 'upstream-health'],
    queryFn: () => adminApi.getUpstreamHealth(),
    refetchInterval: 10000, // Refresh every 10 seconds
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => adminApi.deleteCodexUpstream(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'codex-upstreams'] });
    },
  });

  const toggleStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: 'active' | 'disabled' }) =>
      adminApi.updateCodexUpstreamStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'codex-upstreams'] });
    },
  });

  const triggerHealthCheckMutation = useMutation({
    mutationFn: () => adminApi.triggerHealthCheck(),
    onSuccess: () => {
      setHealthCheckMessage('健康检查已触发，正在检测所有上游...');
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ['admin', 'upstream-health'] });
        setHealthCheckMessage(null);
      }, 2000);
    },
    onError: (error: any) => {
      setHealthCheckMessage(`健康检查失败: ${error.message || '未知错误'}`);
      setTimeout(() => setHealthCheckMessage(null), 3000);
    },
  });

  const upstreams = upstreamsData?.upstreams || [];

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-1 text-xs font-medium text-green-700">
            <CheckCircle2 className="h-3 w-3" />
            正常
          </span>
        );
      case 'unhealthy':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-red-50 px-2 py-1 text-xs font-medium text-red-700">
            <AlertCircle className="h-3 w-3" />
            异常
          </span>
        );
      case 'disabled':
        return (
          <span className="inline-flex items-center gap-1 rounded-full bg-gray-50 px-2 py-1 text-xs font-medium text-gray-700">
            <PowerOff className="h-3 w-3" />
            已禁用
          </span>
        );
      default:
        return null;
    }
  };

  if (isLoading) {
    return <div className="h-64 w-full animate-pulse rounded-xl bg-zinc-100" />;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">Codex 上游管理</h1>
          <p className="text-sm text-zinc-500">管理多个 Codex 上游提供商</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => triggerHealthCheckMutation.mutate()}
            disabled={triggerHealthCheckMutation.isPending}
            className="inline-flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
          >
            <Activity className="h-4 w-4" />
            {triggerHealthCheckMutation.isPending ? '检查中...' : '测试健康'}
          </button>
          <button
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800"
          >
            <Plus className="h-4 w-4" />
            添加上游
          </button>
        </div>
      </div>

      {/* Health Check Message */}
      {healthCheckMessage && (
        <div className="rounded-lg border border-blue-200 bg-blue-50 p-4">
          <div className="flex items-center gap-2">
            <Activity className="h-4 w-4 animate-spin text-blue-600" />
            <p className="text-sm font-medium text-blue-900">{healthCheckMessage}</p>
          </div>
        </div>
      )}

      {/* Upstreams List */}
      <div className="grid gap-4">
        {upstreams.length === 0 ? (
          <div className="rounded-xl border border-zinc-200 bg-white p-12 text-center">
            <Server className="mx-auto h-12 w-12 text-zinc-400" />
            <h3 className="mt-4 text-sm font-medium text-zinc-900">暂无上游配置</h3>
            <p className="mt-2 text-sm text-zinc-500">点击"添加上游"按钮创建第一个上游配置</p>
          </div>
        ) : (
          upstreams.map((upstream) => {
            const health = healthData?.upstreams.find((h) => h.id === upstream.id);
            return (
              <div
                key={upstream.id}
                className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      <Server className="h-5 w-5 text-zinc-400" />
                      <h3 className="text-lg font-semibold text-zinc-900">{upstream.name}</h3>
                      {getStatusBadge(upstream.status)}
                      <span className="rounded-full bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700">
                        优先级 {upstream.priority}
                      </span>
                    </div>
                    <div className="mt-3 grid gap-2 text-sm">
                      <div className="flex items-center gap-2 text-zinc-600">
                        <span className="font-medium">Base URL:</span>
                        <code className="rounded bg-zinc-100 px-2 py-0.5 text-xs">{upstream.base_url}</code>
                      </div>
                      <div className="flex items-center gap-2 text-zinc-600">
                        <span className="font-medium">API Key:</span>
                        <code className="rounded bg-zinc-100 px-2 py-0.5 text-xs">
                          {upstream.api_key.substring(0, 20)}...
                        </code>
                      </div>
                      {health && (
                        <div className="flex items-center gap-4 text-xs text-zinc-500">
                          <span className="flex items-center gap-1">
                            <AlertCircle className="h-3 w-3" />
                            失败次数: {health.failure_count}
                          </span>
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            最后检查: {health.last_checked === 'never' ? '从未' : health.last_checked}
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() =>
                        toggleStatusMutation.mutate({
                          id: upstream.id,
                          status: upstream.status === 'active' ? 'disabled' : 'active',
                        })
                      }
                      className="rounded-lg border border-zinc-200 p-2 text-zinc-600 transition-colors hover:bg-zinc-50"
                      title={upstream.status === 'active' ? '禁用' : '启用'}
                    >
                      {upstream.status === 'active' ? (
                        <PowerOff className="h-4 w-4" />
                      ) : (
                        <Power className="h-4 w-4" />
                      )}
                    </button>
                    <button
                      onClick={() => setEditingUpstream(upstream)}
                      className="rounded-lg border border-zinc-200 p-2 text-zinc-600 transition-colors hover:bg-zinc-50"
                    >
                      <Edit className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => {
                        if (confirm('确定要删除这个上游吗？')) {
                          deleteMutation.mutate(upstream.id);
                        }
                      }}
                      className="rounded-lg border border-red-200 p-2 text-red-600 transition-colors hover:bg-red-50"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Create/Edit Modal */}
      {(showCreateModal || editingUpstream) && (
        <UpstreamModal
          upstream={editingUpstream}
          onClose={() => {
            setShowCreateModal(false);
            setEditingUpstream(null);
          }}
          onSuccess={() => {
            queryClient.invalidateQueries({ queryKey: ['admin', 'codex-upstreams'] });
            setShowCreateModal(false);
            setEditingUpstream(null);
          }}
        />
      )}
    </div>
  );
}

function UpstreamModal({
  upstream,
  onClose,
  onSuccess,
}: {
  upstream?: any;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [formData, setFormData] = useState({
    name: upstream?.name || '',
    base_url: upstream?.base_url || 'https://api.openai.com/v1',
    api_key: upstream?.api_key || '',
    priority: upstream?.priority || 0,
    status: upstream?.status || 'active',
    weight: upstream?.weight || 1,
    max_retries: upstream?.max_retries || 3,
    timeout: upstream?.timeout || 120,
  });

  const mutation = useMutation({
    mutationFn: () =>
      upstream
        ? adminApi.updateCodexUpstream(upstream.id, formData)
        : adminApi.createCodexUpstream(formData),
    onSuccess,
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-2xl rounded-xl bg-white p-6 shadow-xl">
        <h2 className="text-xl font-bold text-zinc-900">{upstream ? '编辑上游' : '添加上游'}</h2>
        <div className="mt-6 space-y-4">
          <div>
            <label className="text-sm font-medium text-zinc-700">名称</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="mt-1 w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
              placeholder="Provider 1"
            />
          </div>
          <div>
            <label className="text-sm font-medium text-zinc-700">Base URL</label>
            <input
              type="text"
              value={formData.base_url}
              onChange={(e) => setFormData({ ...formData, base_url: e.target.value })}
              className="mt-1 w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
            />
          </div>
          <div>
            <label className="text-sm font-medium text-zinc-700">API Key</label>
            <input
              type="password"
              value={formData.api_key}
              onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
              className="mt-1 w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
              placeholder="sk-..."
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-zinc-700">优先级</label>
              <input
                type="number"
                value={formData.priority}
                onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
                className="mt-1 w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
              />
            </div>
            <div>
              <label className="text-sm font-medium text-zinc-700">状态</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                className="mt-1 w-full rounded-lg border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
              >
                <option value="active">启用</option>
                <option value="disabled">禁用</option>
              </select>
            </div>
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="rounded-lg border border-zinc-200 px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
          >
            取消
          </button>
          <button
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending}
            className="rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50"
          >
            {mutation.isPending ? '保存中...' : '保存'}
          </button>
        </div>
      </div>
    </div>
  );
}

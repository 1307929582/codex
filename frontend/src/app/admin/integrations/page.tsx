'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Loader2, Save, ChevronDown, ChevronUp, CheckCircle2, AlertCircle } from 'lucide-react';

interface Provider {
  id: string;
  name: string;
  description: string;
  logo: string;
  fields: {
    api_key: boolean;
    base_url: boolean;
  };
}

const providers: Provider[] = [
  {
    id: 'openai',
    name: 'OpenAI',
    description: 'GPT-4, GPT-3.5 等模型',
    logo: '🤖',
    fields: { api_key: true, base_url: true },
  },
  {
    id: 'anthropic',
    name: 'Anthropic',
    description: 'Claude 系列模型',
    logo: '🧠',
    fields: { api_key: true, base_url: true },
  },
  {
    id: 'google',
    name: 'Google AI',
    description: 'Gemini 系列模型',
    logo: '🔍',
    fields: { api_key: true, base_url: true },
  },
  {
    id: 'azure',
    name: 'Azure OpenAI',
    description: 'Azure 托管的 OpenAI 服务',
    logo: '☁️',
    fields: { api_key: true, base_url: true },
  },
];

export default function AdminIntegrationsPage() {
  const queryClient = useQueryClient();
  const [expandedProvider, setExpandedProvider] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const { data: settings, isLoading } = useQuery({
    queryKey: ['admin', 'settings'],
    queryFn: () => adminApi.getSettings(),
  });

  const [formData, setFormData] = useState({
    openai_api_key: '',
    openai_base_url: 'https://api.openai.com/v1',
  });

  const updateMutation = useMutation({
    mutationFn: () => adminApi.updateSettings(formData),
    onMutate: () => setIsSaving(true),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] });
      setTimeout(() => {
        setIsSaving(false);
        setExpandedProvider(null);
      }, 500);
    },
    onError: () => setIsSaving(false),
  });

  const toggleProvider = (providerId: string) => {
    if (expandedProvider === providerId) {
      setExpandedProvider(null);
    } else {
      setExpandedProvider(providerId);
      if (providerId === 'openai' && settings) {
        setFormData({
          openai_api_key: settings.openai_api_key || '',
          openai_base_url: settings.openai_base_url || 'https://api.openai.com/v1',
        });
      }
    }
  };

  const isConfigured = (providerId: string) => {
    if (providerId === 'openai') {
      return settings?.openai_api_key && settings.openai_api_key.length > 0;
    }
    return false;
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 animate-pulse rounded bg-zinc-200" />
        <div className="grid gap-4 md:grid-cols-2">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-32 animate-pulse rounded-xl bg-zinc-100" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl space-y-6">
      <div className="flex items-center justify-between border-b border-zinc-200 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">服务集���</h1>
          <p className="text-sm text-zinc-500">配置AI服务商API密钥</p>
        </div>
      </div>

      {/* Provider Cards Grid */}
      <div className="grid gap-4 md:grid-cols-2">
        {providers.map((provider) => {
          const configured = isConfigured(provider.id);
          const expanded = expandedProvider === provider.id;

          return (
            <div
              key={provider.id}
              className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-sm transition-all hover:shadow-md"
            >
              {/* Provider Header */}
              <button
                onClick={() => toggleProvider(provider.id)}
                className="flex w-full items-center justify-between p-5 text-left transition-colors hover:bg-zinc-50/50"
              >
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-zinc-100 text-2xl">
                    {provider.logo}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-zinc-900">{provider.name}</h3>
                      {configured ? (
                        <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                      ) : (
                        <AlertCircle className="h-4 w-4 text-zinc-400" />
                      )}
                    </div>
                    <p className="text-sm text-zinc-500">{provider.description}</p>
                  </div>
                </div>
                {expanded ? (
                  <ChevronUp className="h-5 w-5 text-zinc-400" />
                ) : (
                  <ChevronDown className="h-5 w-5 text-zinc-400" />
                )}
              </button>

              {/* Configuration Form */}
              {expanded && (
                <div className="border-t border-zinc-100 bg-zinc-50/30 p-5">
                  {provider.id === 'openai' ? (
                    <div className="space-y-4">
                      <div className="space-y-2">
                        <label className="text-sm font-medium text-zinc-700">API密钥</label>
                        <input
                          type="password"
                          value={formData.openai_api_key}
                          onChange={(e) =>
                            setFormData({ ...formData, openai_api_key: e.target.value })
                          }
                          className="w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm outline-none transition-all focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                          placeholder="sk-..."
                        />
                      </div>
                      <div className="space-y-2">
                        <label className="text-sm font-medium text-zinc-700">Base URL</label>
                        <input
                          type="text"
                          value={formData.openai_base_url}
                          onChange={(e) =>
                            setFormData({ ...formData, openai_base_url: e.target.value })
                          }
                          className="w-full rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm outline-none transition-all focus:border-zinc-900 focus:ring-2 focus:ring-zinc-900/10"
                        />
                      </div>
                      <button
                        onClick={() => updateMutation.mutate()}
                        disabled={isSaving}
                        className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50"
                      >
                        {isSaving ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Save className="h-4 w-4" />
                        )}
                        {isSaving ? '保存中...' : '保存配置'}
                      </button>
                    </div>
                  ) : (
                    <div className="rounded-lg bg-zinc-100 p-4 text-center text-sm text-zinc-500">
                      即将推出
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Info Card */}
      <div className="rounded-xl border border-blue-200 bg-blue-50/50 p-4">
        <div className="flex gap-3">
          <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-blue-100">
            <span className="text-sm">💡</span>
          </div>
          <div className="space-y-1">
            <h4 className="text-sm font-medium text-blue-900">配置提示</h4>
            <p className="text-sm text-blue-700">
              配置API密钥后，系统将能够调用对应服务商的AI模型。请妥善保管您的API密钥，避免泄露。
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

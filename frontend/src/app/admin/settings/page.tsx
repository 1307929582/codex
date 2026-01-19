'use client';

import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Loader2, Save } from 'lucide-react';

export default function AdminSettings() {
  const queryClient = useQueryClient();
  const [isSaving, setIsSaving] = useState(false);

  const { data: settings, isLoading } = useQuery({
    queryKey: ['admin', 'settings'],
    queryFn: () => adminApi.getSettings(),
  });

  const [formData, setFormData] = useState({
    announcement: '',
    default_balance: 0,
    min_recharge_amount: 10,
    registration_enabled: true,
  });

  useEffect(() => {
    if (settings) {
      setFormData({
        announcement: settings.announcement || '',
        default_balance: settings.default_balance || 0,
        min_recharge_amount: settings.min_recharge_amount || 10,
        registration_enabled: settings.registration_enabled ?? true,
      });
    }
  }, [settings]);

  const updateMutation = useMutation({
    mutationFn: () => adminApi.updateSettings(formData),
    onMutate: () => setIsSaving(true),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] });
      setTimeout(() => setIsSaving(false), 500);
    },
    onError: () => setIsSaving(false),
  });

  if (isLoading) return <div className="h-64 w-full animate-pulse rounded-xl bg-zinc-100" />;

  return (
    <div className="max-w-4xl space-y-10">
      <div className="flex items-center justify-between border-b border-zinc-200 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">系统设置</h1>
          <p className="text-sm text-zinc-500">管理全局系统配置</p>
        </div>
        <button
          onClick={() => updateMutation.mutate()}
          disabled={isSaving}
          className="inline-flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50"
        >
          {isSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          {isSaving ? '保存中...' : '保存更改'}
        </button>
      </div>

      {/* User Onboarding */}
      <section className="grid gap-6 md:grid-cols-[200px_1fr]">
        <div>
          <h2 className="text-base font-semibold text-zinc-900">用户入门</h2>
          <p className="text-sm text-zinc-500">新用户的默认设置</p>
        </div>
        <div className="space-y-4 rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
          <div className="grid grid-cols-2 gap-4">
             <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">初始余额 ($)</label>
              <input
                type="number"
                value={formData.default_balance}
                onChange={(e) => setFormData({ ...formData, default_balance: parseFloat(e.target.value) })}
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">最小充值 ($)</label>
              <input
                type="number"
                value={formData.min_recharge_amount}
                onChange={(e) => setFormData({ ...formData, min_recharge_amount: parseFloat(e.target.value) })}
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>
          </div>

          <div className="flex items-center justify-between pt-2">
            <label className="text-sm font-medium text-zinc-700">允许新用户注册</label>
             <button
              type="button"
              onClick={() => setFormData({ ...formData, registration_enabled: !formData.registration_enabled })}
              className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                formData.registration_enabled ? 'bg-zinc-900' : 'bg-zinc-200'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                  formData.registration_enabled ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>
        </div>
      </section>
    </div>
  );
}

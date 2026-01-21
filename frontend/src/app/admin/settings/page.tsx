'use client';

import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { pricingApi } from '@/lib/api/pricing';
import { Loader2, Save, Server, ArrowRight, DollarSign } from 'lucide-react';
import Link from 'next/link';

export default function AdminSettings() {
  const queryClient = useQueryClient();
  const [isSaving, setIsSaving] = useState(false);
  const [globalDailyLimit, setGlobalDailyLimit] = useState('');

  const { data: settings, isLoading } = useQuery({
    queryKey: ['admin', 'settings'],
    queryFn: () => adminApi.getSettings(),
  });

  const { data: pricingData } = useQuery({
    queryKey: ['admin', 'pricing'],
    queryFn: () => pricingApi.list(),
  });

  const pricingList = pricingData?.pricing ?? [];
  const averageMarkup = pricingList.length
    ? (pricingList.reduce((sum, p) => sum + p.markup_multiplier, 0) / pricingList.length).toFixed(2)
    : null;

  const [formData, setFormData] = useState({
    announcement: '',
    default_balance: 0,
    min_recharge_amount: 10,
    email_registration_enabled: false,
    linuxdo_registration_enabled: true,
    openai_api_key: '',
    openai_base_url: '',
    linuxdo_client_id: '',
    linuxdo_client_secret: '',
    linuxdo_enabled: false,
    credit_enabled: false,
    credit_pid: '',
    credit_key: '',
    credit_notify_url: '',
    credit_return_url: '',
    rate_limit_enabled: false,
    rate_limit_rpm: 0,
    rate_limit_burst: 0,
  });

  useEffect(() => {
    if (settings) {
      setFormData({
        announcement: settings.announcement || '',
        default_balance: settings.default_balance || 0,
        min_recharge_amount: settings.min_recharge_amount || 10,
        email_registration_enabled: false,
        linuxdo_registration_enabled: settings.linuxdo_registration_enabled ?? true,
        openai_api_key: settings.openai_api_key || '',
        openai_base_url: settings.openai_base_url || '',
        linuxdo_client_id: settings.linuxdo_client_id || '',
        linuxdo_client_secret: settings.linuxdo_client_secret || '',
        linuxdo_enabled: settings.linuxdo_enabled ?? false,
        credit_enabled: settings.credit_enabled ?? false,
        credit_pid: settings.credit_pid || '',
        credit_key: settings.credit_key || '',
        credit_notify_url: settings.credit_notify_url || '',
        credit_return_url: settings.credit_return_url || '',
        rate_limit_enabled: settings.rate_limit_enabled ?? false,
        rate_limit_rpm: settings.rate_limit_rpm ?? 0,
        rate_limit_burst: settings.rate_limit_burst ?? 0,
      });
      setGlobalDailyLimit(
        settings.user_daily_usage_limit === null || settings.user_daily_usage_limit === undefined
          ? ''
          : settings.user_daily_usage_limit.toString()
      );
    }
  }, [settings]);

  const updateMutation = useMutation({
    mutationFn: (payload: any) => adminApi.updateSettings(payload),
    onMutate: () => setIsSaving(true),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] });
      setTimeout(() => setIsSaving(false), 500);
      alert('设置保存成功！');
    },
    onError: (error: any) => {
      setIsSaving(false);
      alert('保存失败，请稍后重试');
      console.error('Settings update error:', error);
    },
  });

  const handleSave = () => {
    const value = globalDailyLimit.trim();
    if (value !== '') {
      const parsed = Number(value);
      if (Number.isNaN(parsed) || parsed < 0) {
        alert('请输入有效的每日上限');
        return;
      }
    }

    const parsedLimit = value === '' ? null : Number(value);

    updateMutation.mutate({
      ...formData,
      user_daily_usage_limit: parsedLimit,
    });
  };

  if (isLoading) return <div className="h-64 w-full animate-pulse rounded-xl bg-zinc-100" />;

  return (
    <div className="max-w-4xl space-y-10">
      <div className="flex items-center justify-between border-b border-zinc-200 pb-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900">系统设置</h1>
          <p className="text-sm text-zinc-500">管理全局系统配置</p>
        </div>
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="inline-flex items-center gap-2 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50"
        >
          {isSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          {isSaving ? '保存中...' : '保存更改'}
        </button>
      </div>

      {/* Codex Configuration Notice */}
      <section className="rounded-xl border border-blue-200 bg-blue-50 p-6">
        <div className="flex items-start gap-4">
          <div className="rounded-lg bg-blue-100 p-2">
            <Server className="h-5 w-5 text-blue-600" />
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-blue-900">Codex 上游配置已迁移</h3>
            <p className="mt-1 text-sm text-blue-700">
              Codex 配置现已移至专用的上游管理页面，支持多上游提供商、健康检查和用户会话亲和性。
            </p>
            <Link
              href="/admin/upstreams"
              className="mt-3 inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
            >
              前往 Codex 上游管理
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* Pricing Management Notice */}
      <section className="rounded-xl border border-emerald-200 bg-emerald-50 p-6">
        <div className="flex items-start gap-4">
          <div className="rounded-lg bg-emerald-100 p-2">
            <DollarSign className="h-5 w-5 text-emerald-600" />
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-emerald-900">定价与价格比例管理</h3>
            <p className="mt-1 text-sm text-emerald-700">
              管理所有模型的定价和价格比例（利润率）。当前平均价格比例：
              <span className="ml-1 font-semibold">
                {averageMarkup ? `${averageMarkup}x` : '加载中...'}
              </span>
            </p>
            <Link
              href="/admin/pricing"
              className="mt-3 inline-flex items-center gap-2 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-emerald-700"
            >
              前往定价管理
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* User Onboarding */}
      <section className="grid gap-6 md:grid-cols-[200px_1fr] pt-6 border-t border-zinc-100">
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
            <label className="text-sm font-medium text-zinc-700">允许 LinuxDo 注册</label>
             <button
              type="button"
              onClick={() => setFormData({ ...formData, linuxdo_registration_enabled: !formData.linuxdo_registration_enabled })}
              className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                formData.linuxdo_registration_enabled ? 'bg-zinc-900' : 'bg-zinc-200'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                  formData.linuxdo_registration_enabled ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>
        </div>
      </section>

      {/* Usage Limit */}
      <section className="grid gap-6 md:grid-cols-[200px_1fr] pt-6 border-t border-zinc-100">
        <div>
          <h2 className="text-base font-semibold text-zinc-900">用量限制</h2>
          <p className="text-sm text-zinc-500">对所有用户统一生效</p>
        </div>
        <div className="space-y-4 rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
          <div className="space-y-2">
            <label className="text-sm font-medium text-zinc-700">用户每日最大使用量 ($)</label>
            <input
              type="number"
              step="0.01"
              value={globalDailyLimit}
              onChange={(e) => setGlobalDailyLimit(e.target.value)}
              placeholder="留空表示不限制"
              className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
            />
            <p className="text-xs text-zinc-500">留空表示不限制，0 表示禁止使用</p>
          </div>
        </div>
      </section>

      {/* LinuxDo OAuth Configuration */}
      <section className="grid gap-6 md:grid-cols-[200px_1fr] pt-6 border-t border-zinc-100">
        <div>
          <h2 className="text-base font-semibold text-zinc-900">LinuxDo OAuth</h2>
          <p className="text-sm text-zinc-500">LinuxDo 社区登录配置</p>
        </div>
        <div className="space-y-4 rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between pb-4 border-b border-zinc-100">
            <div>
              <label className="text-sm font-medium text-zinc-700">启用 LinuxDo 登录</label>
              <p className="text-xs text-zinc-500 mt-1">允许用户使用 LinuxDo 账户登录</p>
            </div>
            <button
              type="button"
              onClick={() => setFormData({ ...formData, linuxdo_enabled: !formData.linuxdo_enabled })}
              className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                formData.linuxdo_enabled ? 'bg-zinc-900' : 'bg-zinc-200'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                  formData.linuxdo_enabled ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">Client ID</label>
              <input
                type="text"
                value={formData.linuxdo_client_id}
                onChange={(e) => setFormData({ ...formData, linuxdo_client_id: e.target.value })}
                placeholder="kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ"
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm font-mono outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">Client Secret</label>
              <input
                type="password"
                value={formData.linuxdo_client_secret}
                onChange={(e) => setFormData({ ...formData, linuxdo_client_secret: e.target.value })}
                placeholder="••••••••••••••••••••••••••••••••"
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm font-mono outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>

            <div className="rounded-lg bg-blue-50 p-4 text-sm text-blue-700">
              <p className="font-medium mb-1">回调地址配置</p>
              <p className="text-xs">在 LinuxDo 应用设置中，请将回调地址设置为：</p>
              <code className="block mt-2 rounded bg-blue-100 px-2 py-1 text-xs font-mono text-blue-900">
                https://codex.zenscaleai.com/api/auth/linuxdo/callback
              </code>
            </div>
          </div>
        </div>
      </section>

      {/* Rate Limit Configuration */}
      <section className="grid gap-6 md:grid-cols-[200px_1fr] pt-6 border-t border-zinc-100">
        <div>
          <h2 className="text-base font-semibold text-zinc-900">限流设置</h2>
          <p className="text-sm text-zinc-500">按 API Key 限制请求速率</p>
        </div>
        <div className="space-y-4 rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between pb-4 border-b border-zinc-100">
            <div>
              <label className="text-sm font-medium text-zinc-700">启用限流</label>
              <p className="text-xs text-zinc-500 mt-1">限制每个 API Key 的请求速率</p>
            </div>
            <button
              type="button"
              onClick={() => setFormData({ ...formData, rate_limit_enabled: !formData.rate_limit_enabled })}
              className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                formData.rate_limit_enabled ? 'bg-zinc-900' : 'bg-zinc-200'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                  formData.rate_limit_enabled ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">每分钟请求数</label>
              <input
                type="number"
                value={formData.rate_limit_rpm}
                onChange={(e) => setFormData({ ...formData, rate_limit_rpm: parseInt(e.target.value || '0', 10) })}
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">突发额度</label>
              <input
                type="number"
                value={formData.rate_limit_burst}
                onChange={(e) => setFormData({ ...formData, rate_limit_burst: parseInt(e.target.value || '0', 10) })}
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>
          </div>

          <div className="rounded-lg bg-zinc-50 p-4 text-xs text-zinc-600">
            <p>提示：当突发额度为 0 时，将自动使用「每分钟请求数」作为默认突发值。</p>
          </div>
        </div>
      </section>

      {/* Credit Payment Configuration */}
      <section className="grid gap-6 md:grid-cols-[200px_1fr] pt-6 border-t border-zinc-100">
        <div>
          <h2 className="text-base font-semibold text-zinc-900">Credit 支付</h2>
          <p className="text-sm text-zinc-500">Linux Do Credit 支付配置</p>
        </div>
        <div className="space-y-4 rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
          <div className="flex items-center justify-between pb-4 border-b border-zinc-100">
            <div>
              <label className="text-sm font-medium text-zinc-700">启用 Credit 支付</label>
              <p className="text-xs text-zinc-500 mt-1">允许用户使用 Linux Do Credit 购买套餐</p>
            </div>
            <button
              type="button"
              onClick={() => setFormData({ ...formData, credit_enabled: !formData.credit_enabled })}
              className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                formData.credit_enabled ? 'bg-zinc-900' : 'bg-zinc-200'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                  formData.credit_enabled ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>

          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">PID (Client ID)</label>
              <input
                type="text"
                value={formData.credit_pid}
                onChange={(e) => setFormData({ ...formData, credit_pid: e.target.value })}
                placeholder="001"
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm font-mono outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">Key (Client Secret)</label>
              <input
                type="password"
                value={formData.credit_key}
                onChange={(e) => setFormData({ ...formData, credit_key: e.target.value })}
                placeholder="••••••••••••••••••••••••••••••••"
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm font-mono outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">Notify URL (异步通知地址)</label>
              <input
                type="text"
                value={formData.credit_notify_url}
                onChange={(e) => setFormData({ ...formData, credit_notify_url: e.target.value })}
                placeholder="https://your-domain.com/api/payment/credit/notify"
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm font-mono outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-zinc-700">Return URL (同步返回地址)</label>
              <input
                type="text"
                value={formData.credit_return_url}
                onChange={(e) => setFormData({ ...formData, credit_return_url: e.target.value })}
                placeholder="https://your-domain.com/packages"
                className="w-full rounded-md border border-zinc-200 px-3 py-2 text-sm font-mono outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/10"
              />
            </div>

            <div className="rounded-lg bg-amber-50 p-4 text-sm text-amber-700">
              <p className="font-medium mb-1">重要提示</p>
              <ul className="text-xs space-y-1 list-disc list-inside">
                <li>Notify URL 必须是外网可访问的 HTTPS 地址</li>
                <li>在 Linux Do Credit 平台创建应用时，需要配置这些回调地址</li>
                <li>PID 和 Key 请妥善保管，不要泄露</li>
              </ul>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}

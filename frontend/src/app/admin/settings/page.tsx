'use client';

import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Save, Key, Users, Bell, CheckCircle2, AlertCircle } from 'lucide-react';

export default function AdminSettingsPage() {
  const queryClient = useQueryClient();
  const [showSuccess, setShowSuccess] = useState(false);

  const { data: settings, isLoading } = useQuery({
    queryKey: ['admin', 'settings'],
    queryFn: () => adminApi.getSettings(),
  });

  const [formData, setFormData] = useState({
    announcement: '',
    default_balance: 0,
    min_recharge_amount: 10,
    registration_enabled: true,
    openai_api_key: '',
    openai_base_url: 'https://api.openai.com/v1',
  });

  useEffect(() => {
    if (settings) {
      setFormData({
        announcement: settings.announcement || '',
        default_balance: settings.default_balance || 0,
        min_recharge_amount: settings.min_recharge_amount || 10,
        registration_enabled: settings.registration_enabled ?? true,
        openai_api_key: settings.openai_api_key || '',
        openai_base_url: settings.openai_base_url || 'https://api.openai.com/v1',
      });
    }
  }, [settings]);

  const updateMutation = useMutation({
    mutationFn: () => adminApi.updateSettings(formData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] });
      setShowSuccess(true);
      setTimeout(() => setShowSuccess(false), 3000);
    },
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">系统设置</h1>
          <p className="mt-2 text-sm text-gray-600">配置系统参数和服务集成</p>
        </div>
        <button
          onClick={() => updateMutation.mutate()}
          disabled={updateMutation.isPending}
          className="flex items-center gap-2 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-3 text-white shadow-lg shadow-blue-500/30 transition-all hover:shadow-xl hover:shadow-blue-500/40 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Save className="h-5 w-5" />
          {updateMutation.isPending ? '保存中...' : '保存设置'}
        </button>
      </div>

      {/* Success Message */}
      {showSuccess && (
        <div className="rounded-xl bg-green-50 border border-green-200 p-4 flex items-center gap-3 animate-in fade-in slide-in-from-top-2">
          <CheckCircle2 className="h-5 w-5 text-green-600" />
          <p className="text-sm font-medium text-green-900">设置已成功保存</p>
        </div>
      )}

      {/* OpenAI Configuration */}
      <div className="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-900/5">
        <div className="flex items-center gap-3 mb-6">
          <div className="rounded-lg bg-gradient-to-br from-purple-500 to-purple-600 p-2.5 shadow-lg">
            <Key className="h-5 w-5 text-white" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">OpenAI 配置</h2>
            <p className="text-sm text-gray-600">配置 OpenAI API 密钥和服务端点</p>
          </div>
        </div>

        <div className="space-y-5">
          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              API 密钥
            </label>
            <div className="relative">
              <input
                type="password"
                value={formData.openai_api_key}
                onChange={(e) =>
                  setFormData({ ...formData, openai_api_key: e.target.value })
                }
                className="w-full rounded-xl border border-gray-300 px-4 py-3 pr-10 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all"
                placeholder="sk-..."
              />
              <Key className="absolute right-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
            </div>
            <p className="mt-2 text-sm text-gray-500 flex items-start gap-2">
              <AlertCircle className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <span>用于调用 OpenAI 服务的 API 密钥，请妥善保管</span>
            </p>
          </div>

          <div>
            <label className="block text-sm font-semibold text-gray-700 mb-2">
              Base URL
            </label>
            <input
              type="text"
              value={formData.openai_base_url}
              onChange={(e) =>
                setFormData({ ...formData, openai_base_url: e.target.value })
              }
              className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all"
              placeholder="https://api.openai.com/v1"
            />
            <p className="mt-2 text-sm text-gray-500">
              OpenAI API 的 Base URL，支持自定义代理服务
            </p>
          </div>
        </div>
      </div>

      {/* System Announcement */}
      <div className="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-900/5">
        <div className="flex items-center gap-3 mb-6">
          <div className="rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 p-2.5 shadow-lg">
            <Bell className="h-5 w-5 text-white" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">系统公告</h2>
            <p className="text-sm text-gray-600">设置显示在用户控制台的公告信息</p>
          </div>
        </div>

        <div>
          <label className="block text-sm font-semibold text-gray-700 mb-2">
            公告内容
          </label>
          <textarea
            value={formData.announcement}
            onChange={(e) =>
              setFormData({ ...formData, announcement: e.target.value })
            }
            className="w-full rounded-xl border border-gray-300 px-4 py-3 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all resize-none"
            rows={4}
            placeholder="在此输入系统公告，将显示在用户控制台..."
          />
          <p className="mt-2 text-sm text-gray-500">
            支持 Markdown 格式，留空则不显示公告
          </p>
        </div>
      </div>

      {/* User Settings */}
      <div className="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-900/5">
        <div className="flex items-center gap-3 mb-6">
          <div className="rounded-lg bg-gradient-to-br from-green-500 to-emerald-600 p-2.5 shadow-lg">
            <Users className="h-5 w-5 text-white" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-gray-900">用户设置</h2>
            <p className="text-sm text-gray-600">配置新用户注册和默认参数</p>
          </div>
        </div>

        <div className="space-y-5">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                新用户默认余额
              </label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 font-medium">$</span>
                <input
                  type="number"
                  step="0.01"
                  value={formData.default_balance}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      default_balance: parseFloat(e.target.value) || 0,
                    })
                  }
                  className="w-full rounded-xl border border-gray-300 pl-8 pr-4 py-3 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all"
                />
              </div>
              <p className="mt-2 text-sm text-gray-500">
                新注册用户将获得此金额的初始余额
              </p>
            </div>

            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                最小充值金额
              </label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 font-medium">$</span>
                <input
                  type="number"
                  step="0.01"
                  value={formData.min_recharge_amount}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      min_recharge_amount: parseFloat(e.target.value) || 10,
                    })
                  }
                  className="w-full rounded-xl border border-gray-300 pl-8 pr-4 py-3 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 focus:outline-none transition-all"
                />
              </div>
              <p className="mt-2 text-sm text-gray-500">
                用户单次充值的最小金额限制
              </p>
            </div>
          </div>

          <div className="flex items-center justify-between p-4 rounded-xl bg-gray-50 border border-gray-200">
            <div className="flex items-center gap-3">
              <div className="flex-shrink-0">
                <Users className="h-5 w-5 text-gray-600" />
              </div>
              <div>
                <p className="text-sm font-semibold text-gray-900">允许新用户注册</p>
                <p className="text-sm text-gray-500">关闭后将禁止新用户注册账户</p>
              </div>
            </div>
            <button
              type="button"
              onClick={() =>
                setFormData({
                  ...formData,
                  registration_enabled: !formData.registration_enabled,
                })
              }
              className={`relative inline-flex h-7 w-12 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 ${
                formData.registration_enabled ? 'bg-blue-600' : 'bg-gray-300'
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-6 w-6 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                  formData.registration_enabled ? 'translate-x-5' : 'translate-x-0'
                }`}
              />
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { adminApi } from '@/lib/api/admin';
import { Save } from 'lucide-react';

export default function AdminSettingsPage() {
  const queryClient = useQueryClient();

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

  // Update form when data loads
  useState(() => {
    if (settings) {
      setFormData({
        announcement: settings.announcement || '',
        default_balance: settings.default_balance || 0,
        min_recharge_amount: settings.min_recharge_amount || 10,
        registration_enabled: settings.registration_enabled ?? true,
      });
    }
  });

  const updateMutation = useMutation({
    mutationFn: () => adminApi.updateSettings(formData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] });
      alert('设置已保存');
    },
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8 flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">系统设置</h1>
        <button
          onClick={() => updateMutation.mutate()}
          disabled={updateMutation.isPending}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
        >
          <Save className="h-5 w-5" />
          {updateMutation.isPending ? '保存中...' : '保存设置'}
        </button>
      </div>

      <div className="space-y-6">
        <div className="rounded-lg bg-white p-6 shadow-sm">
          <h2 className="mb-4 text-xl font-bold text-gray-900">系统公告</h2>
          <textarea
            value={formData.announcement}
            onChange={(e) =>
              setFormData({ ...formData, announcement: e.target.value })
            }
            className="w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
            rows={4}
            placeholder="在此输入系统公告，将显示在用户Dashboard..."
          />
        </div>

        <div className="rounded-lg bg-white p-6 shadow-sm">
          <h2 className="mb-4 text-xl font-bold text-gray-900">用户设置</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                新用户默认余额
              </label>
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
                className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
              />
              <p className="mt-1 text-sm text-gray-500">
                新注册用户将获得此金额的初始余额
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700">
                最小充值金额
              </label>
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
                className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
              />
              <p className="mt-1 text-sm text-gray-500">
                用户单次充值的最小金额限制
              </p>
            </div>

            <div className="flex items-center">
              <input
                type="checkbox"
                id="registration"
                checked={formData.registration_enabled}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    registration_enabled: e.target.checked,
                  })
                }
                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <label
                htmlFor="registration"
                className="ml-2 block text-sm text-gray-900"
              >
                允许新用户注册
              </label>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

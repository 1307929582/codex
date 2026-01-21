'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { useAuthStore } from '@/lib/stores/auth';
import { Loader2 } from 'lucide-react';

export default function SetupWizard() {
  const router = useRouter();
  const { setAuth } = useAuthStore();
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(true);
  const [needsSetup, setNeedsSetup] = useState(false);

  const [formData, setFormData] = useState({
    // Admin account
    email: '',
    password: '',
    confirmPassword: '',
    // OpenAI config
    openai_api_key: '',
    openai_base_url: 'https://api.openai.com/v1',
    // System settings
    announcement: '',
    default_balance: 0,
    email_registration_enabled: false,
    linuxdo_registration_enabled: true,
  });

  // Check if setup is needed
  useEffect(() => {
    const checkSetup = async () => {
      try {
        const response = await api.get('/api/setup/status');
        if (response.data.needs_setup) {
          setNeedsSetup(true);
        } else {
          router.push('/');
        }
      } catch (error) {
        setNeedsSetup(true);
      } finally {
        setLoading(false);
      }
    };
    checkSetup();
  }, [router]);

  const setupMutation = useMutation({
    mutationFn: async () => {
      const response = await api.post('/api/setup/initialize', formData);
      return response.data;
    },
    onSuccess: async (data) => {
      // Auto login with the created admin account
      localStorage.setItem('token', data.token);

      // Fetch user info and save to auth store
      try {
        const userResponse = await api.get('/api/auth/me');
        setAuth(data.token, userResponse.data);
      } catch (error) {
        console.error('Failed to fetch user info:', error);
      }

      router.push('/admin');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (step === 1) {
      // Validate admin account
      if (!formData.email || !formData.password) {
        alert('请填写邮箱和密码');
        return;
      }
      if (formData.password !== formData.confirmPassword) {
        alert('两次密码输入不一致');
        return;
      }
      if (formData.password.length < 6) {
        alert('密码至少6个字符');
        return;
      }
      setStep(2);
    } else if (step === 2) {
      // Validate OpenAI config
      if (!formData.openai_api_key) {
        alert('请输入OpenAI API密钥');
        return;
      }
      setStep(3);
    } else {
      // Submit
      setupMutation.mutate();
    }
  };

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  if (!needsSetup) {
    return null;
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-100 px-4">
      <div className="w-full max-w-2xl rounded-lg bg-white p-8 shadow-lg">
        {/* Header */}
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-gray-900">
            欢迎使用 Zenscale Codex
          </h1>
          <p className="mt-2 text-gray-600">
            首次安装向导 - 第 {step} / 3 步
          </p>
        </div>

        {/* Progress bar */}
        <div className="mb-8">
          <div className="flex justify-between">
            {[1, 2, 3].map((s) => (
              <div
                key={s}
                className={`h-2 flex-1 ${s < 3 ? 'mr-2' : ''} rounded ${
                  s <= step ? 'bg-blue-600' : 'bg-gray-200'
                }`}
              />
            ))}
          </div>
          <div className="mt-2 flex justify-between text-sm text-gray-600">
            <span>创建管理员</span>
            <span>配置OpenAI</span>
            <span>系统设置</span>
          </div>
        </div>

        <form onSubmit={handleSubmit}>
          {/* Step 1: Admin Account */}
          {step === 1 && (
            <div className="space-y-4">
              <h2 className="text-xl font-bold text-gray-900">
                创建管理员账户
              </h2>
              <p className="text-sm text-gray-600">
                创建第一个管理员账户，用于管理系统
              </p>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  邮箱地址
                </label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) =>
                    setFormData({ ...formData, email: e.target.value })
                  }
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  placeholder="admin@example.com"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  密码
                </label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) =>
                    setFormData({ ...formData, password: e.target.value })
                  }
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  placeholder="至少6个字符"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  确认密码
                </label>
                <input
                  type="password"
                  value={formData.confirmPassword}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      confirmPassword: e.target.value,
                    })
                  }
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  placeholder="再次输入密码"
                  required
                />
              </div>
            </div>
          )}

          {/* Step 2: OpenAI Config */}
          {step === 2 && (
            <div className="space-y-4">
              <h2 className="text-xl font-bold text-gray-900">配置OpenAI</h2>
              <p className="text-sm text-gray-600">
                配置OpenAI API密钥，用于调用AI服务
              </p>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  OpenAI API密钥
                </label>
                <input
                  type="password"
                  value={formData.openai_api_key}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      openai_api_key: e.target.value,
                    })
                  }
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  placeholder="sk-..."
                  required
                />
                <p className="mt-1 text-sm text-gray-500">
                  从 OpenAI 官网获取：
                  <a
                    href="https://platform.openai.com/api-keys"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 hover:underline"
                  >
                    platform.openai.com/api-keys
                  </a>
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  OpenAI Base URL（可选）
                </label>
                <input
                  type="text"
                  value={formData.openai_base_url}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      openai_base_url: e.target.value,
                    })
                  }
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  placeholder="https://api.openai.com/v1"
                />
                <p className="mt-1 text-sm text-gray-500">
                  如果使用代理，请修改此URL
                </p>
              </div>
            </div>
          )}

          {/* Step 3: System Settings */}
          {step === 3 && (
            <div className="space-y-4">
              <h2 className="text-xl font-bold text-gray-900">系统设置</h2>
              <p className="text-sm text-gray-600">
                配置系统参数（可选，稍后可在管理面板修改）
              </p>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  系统公告（可选）
                </label>
                <textarea
                  value={formData.announcement}
                  onChange={(e) =>
                    setFormData({ ...formData, announcement: e.target.value })
                  }
                  className="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-blue-500 focus:outline-none"
                  rows={3}
                  placeholder="欢迎使用 Zenscale Codex..."
                />
              </div>

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
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="linuxdo-registration"
                  checked={formData.linuxdo_registration_enabled}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      linuxdo_registration_enabled: e.target.checked,
                    })
                  }
                  className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label
                  htmlFor="linuxdo-registration"
                  className="ml-2 block text-sm text-gray-900"
                >
                  允许 LinuxDo 注册
                </label>
              </div>
            </div>
          )}

          {/* Buttons */}
          <div className="mt-8 flex justify-between">
            {step > 1 && (
              <button
                type="button"
                onClick={() => setStep(step - 1)}
                className="rounded-lg border border-gray-300 px-6 py-2 text-gray-700 hover:bg-gray-50"
              >
                上一步
              </button>
            )}
            <button
              type="submit"
              disabled={setupMutation.isPending}
              className="ml-auto rounded-lg bg-blue-600 px-6 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {setupMutation.isPending
                ? '初始化中...'
                : step === 3
                ? '完成设置'
                : '下一步'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

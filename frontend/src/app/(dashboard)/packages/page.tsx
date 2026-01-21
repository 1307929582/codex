'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { packageApi } from '@/lib/api/package';
import { Check, Zap, Clock, DollarSign } from 'lucide-react';

export default function PackagesPage() {
  const [purchasing, setPurchasing] = useState<number | null>(null);
  const [couponCode, setCouponCode] = useState('');

  const { data, isLoading } = useQuery({
    queryKey: ['packages'],
    queryFn: () => packageApi.list(),
  });

  const { data: dailyUsage } = useQuery({
    queryKey: ['daily-usage'],
    queryFn: () => packageApi.getDailyUsage(),
  });

  const activePackage = dailyUsage?.package;

  const submitPayment = (result: { payment_url?: string; params?: Record<string, string> }) => {
    if (!result.payment_url || !result.params) {
      alert('操作成功');
      setPurchasing(null);
      return;
    }

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
  };

  const handlePurchase = async (packageId: number) => {
    try {
      setPurchasing(packageId);
      const result = await packageApi.purchase(packageId, couponCode.trim() || undefined);
      submitPayment(result);
    } catch (error) {
      console.error('Purchase failed:', error);
      alert('购买失败，请稍后重试');
      setPurchasing(null);
    }
  };

  const handleSwitch = async (packageId: number) => {
    try {
      setPurchasing(packageId);
      const result = await packageApi.switchPackage(packageId, couponCode.trim() || undefined);
      if (!result.payment_url) {
        const creditText = result.balance_credit ? `，余额补偿 $${result.balance_credit.toFixed(2)}` : '';
        alert(`套餐已切换${creditText}`);
        setPurchasing(null);
        return;
      }
      submitPayment(result);
    } catch (error) {
      console.error('Switch failed:', error);
      alert('切换失败，请稍后重试');
      setPurchasing(null);
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto space-y-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold tracking-tight text-zinc-900">选择套餐</h1>
        <p className="mt-2 text-zinc-600">选择适合您的套餐，享受每日额度</p>
      </div>

      <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-zinc-900">优惠码</h2>
            <p className="text-sm text-zinc-500">购买或切换套餐时可使用</p>
          </div>
          <input
            value={couponCode}
            onChange={(e) => setCouponCode(e.target.value)}
            placeholder="输入优惠码"
            className="w-full rounded-lg border border-zinc-200 px-4 py-2 text-sm md:w-72"
          />
        </div>
      </div>

      {activePackage && (
        <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-6">
          <h3 className="text-sm font-semibold text-emerald-900">当前套餐</h3>
          <div className="mt-2 flex flex-col gap-1 text-sm text-emerald-800">
            <span>{activePackage.package_name}</span>
            <span>有效期至 {new Date(activePackage.end_date).toLocaleDateString('zh-CN')}</span>
            <span>切换套餐将按剩余天数折算</span>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {data?.packages.map((pkg) => (
          <div
            key={pkg.id}
            className="relative rounded-2xl border-2 border-zinc-200 bg-white p-8 shadow-sm transition-all hover:border-zinc-900 hover:shadow-lg"
          >
            {/* Package Header */}
            <div className="mb-6">
              <h3 className="text-2xl font-bold text-zinc-900">{pkg.name}</h3>
              <p className="mt-2 text-sm text-zinc-600">{pkg.description}</p>
            </div>

            {/* Price */}
            <div className="mb-6">
              <div className="flex items-baseline">
                <span className="text-4xl font-bold text-zinc-900">${pkg.price.toFixed(2)}</span>
                <span className="ml-2 text-zinc-600">/ {pkg.duration_days}天</span>
              </div>
            </div>

            {/* Features */}
            <ul className="mb-8 space-y-3">
              <li className="flex items-center gap-3 text-sm text-zinc-700">
                <div className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100">
                  <Check className="h-3 w-3 text-emerald-600" />
                </div>
                <span>每日限额 ${pkg.daily_limit.toFixed(2)}</span>
              </li>
              <li className="flex items-center gap-3 text-sm text-zinc-700">
                <div className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100">
                  <Clock className="h-3 w-3 text-emerald-600" />
                </div>
                <span>有效期 {pkg.duration_days} 天</span>
              </li>
              <li className="flex items-center gap-3 text-sm text-zinc-700">
                <div className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100">
                  <Zap className="h-3 w-3 text-emerald-600" />
                </div>
                <span>每日自动重置</span>
              </li>
              <li className="flex items-center gap-3 text-sm text-zinc-700">
                <div className="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100">
                  <DollarSign className="h-3 w-3 text-emerald-600" />
                </div>
                <span>超额自动从余额扣费</span>
              </li>
            </ul>

            {/* Stock Info */}
            {pkg.stock !== -1 && (
              <div className="mb-4 rounded-lg bg-zinc-50 px-4 py-2 text-center text-sm">
                <span className={pkg.stock > 0 ? 'text-zinc-600' : 'text-red-600'}>
                  {pkg.stock > 0 ? `剩余 ${pkg.stock} 份` : '已售罄'}
                </span>
              </div>
            )}

            {/* Purchase Button */}
            <button
              onClick={() => (activePackage && activePackage.package_id !== pkg.id ? handleSwitch(pkg.id) : handlePurchase(pkg.id))}
              disabled={
                purchasing === pkg.id ||
                (pkg.stock !== -1 && pkg.stock <= 0) ||
                (activePackage && activePackage.package_id === pkg.id)
              }
              className="w-full rounded-lg bg-zinc-900 px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {purchasing === pkg.id
                ? '处理中...'
                : pkg.stock !== -1 && pkg.stock <= 0
                  ? '已售罄'
                  : activePackage && activePackage.package_id === pkg.id
                    ? '当前套餐'
                    : activePackage
                      ? '切换套餐'
                      : '立即购买'}
            </button>
          </div>
        ))}
      </div>

      {/* Info Section */}
      <div className="rounded-xl border border-blue-200 bg-blue-50 p-6">
        <h3 className="text-sm font-semibold text-blue-900">套餐说明</h3>
        <ul className="mt-3 space-y-2 text-sm text-blue-700">
          <li>• 套餐购买后立即生效，有效期从购买日开始计算</li>
          <li>• 每日限额在UTC+8时区的0点自动重置</li>
          <li>• 当日额度用完后，会自动从账户余额扣费</li>
          <li>• 套餐到期后自动失效，切换回按量付费模式</li>
          <li>• 支持使用 Linux Do Credit 支付</li>
        </ul>
      </div>
    </div>
  );
}

'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuthStore } from '@/lib/stores/auth';
import {
  LayoutDashboard,
  Users,
  Settings,
  Activity,
  LogOut,
  ChevronRight,
  ShieldCheck,
  Server,
  Package,
  Receipt,
  FileText
} from 'lucide-react';

const navigation = [
  { name: '控制台', href: '/admin', icon: LayoutDashboard },
  { name: '用户管理', href: '/admin/users', icon: Users },
  { name: '套餐管理', href: '/admin/packages', icon: Package },
  { name: '订单管理', href: '/admin/orders', icon: Receipt },
  { name: '使用记录', href: '/admin/usage', icon: FileText },
  { name: 'Codex 上游', href: '/admin/upstreams', icon: Server },
  { name: '系统设置', href: '/admin/settings', icon: Settings },
  { name: '操作日志', href: '/admin/logs', icon: Activity },
];

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const { user, logout } = useAuthStore();
  const [isLoading, setIsLoading] = useState(true);
  const displayName = user ? user.username || '未设置' : '未登录';
  const linuxdoId = user?.oauth_provider === 'linuxdo' ? user?.oauth_id : '';
  const avatarInitial = (displayName || 'U')[0].toUpperCase();

  useEffect(() => {
    setIsLoading(false);
  }, []);

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-zinc-900 mx-auto"></div>
          <p className="mt-4 text-zinc-600">加载中...</p>
        </div>
      </div>
    );
  }

  if (!user || (user.role !== 'admin' && user.role !== 'super_admin')) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50">
        <div className="w-full max-w-md space-y-8 rounded-2xl bg-white p-10 text-center shadow-xl ring-1 ring-zinc-900/5">
          <ShieldCheck className="mx-auto h-12 w-12 text-red-500" />
          <h1 className="mt-4 text-2xl font-bold tracking-tight text-zinc-900">访问被拒绝</h1>
          <p className="text-zinc-500">您没有权限访问管理员面板</p>
          <p className="mt-2 text-sm text-zinc-500">当前用户: {displayName || '未登录'}</p>
          <p className="mt-1 text-sm text-zinc-500">
            LinuxDo ID: {linuxdoId || '未绑定'}
          </p>
          <p className="mt-1 text-sm text-zinc-500">用户角色: {user?.role || '无'}</p>
          <Link href="/" className="inline-block text-sm font-medium text-blue-600 hover:text-blue-500">
            返回首页 &rarr;
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-zinc-50/50">
      {/* Sidebar */}
      <aside className="fixed inset-y-0 left-0 z-50 w-64 border-r border-zinc-200 bg-white/80 backdrop-blur-xl">
        <div className="flex h-16 items-center px-6 border-b border-zinc-100">
          <div className="flex items-center gap-2 font-bold text-zinc-900">
            <div className="h-6 w-6 rounded-md bg-zinc-900" />
            <span>Zenscale Codex</span>
          </div>
        </div>

        <nav className="flex flex-col gap-1 p-4">
          <div className="px-2 text-xs font-medium text-zinc-400 uppercase tracking-wider mb-2">平台管理</div>
          {navigation.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.name}
                href={item.href}
                className={`group flex items-center justify-between rounded-lg px-3 py-2 text-sm font-medium transition-all duration-200 ${
                  isActive
                    ? 'bg-zinc-100 text-zinc-900'
                    : 'text-zinc-500 hover:bg-zinc-50 hover:text-zinc-900'
                }`}
              >
                <div className="flex items-center gap-3">
                  <item.icon className={`h-4 w-4 ${isActive ? 'text-zinc-900' : 'text-zinc-400 group-hover:text-zinc-900'}`} />
                  {item.name}
                </div>
                {isActive && <ChevronRight className="h-3 w-3 text-zinc-400" />}
              </Link>
            );
          })}
        </nav>

        <div className="absolute bottom-0 w-full border-t border-zinc-100 bg-zinc-50/50 p-4">
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-zinc-200 text-xs font-medium text-zinc-600">
              {avatarInitial}
            </div>
            <div className="flex-1 overflow-hidden">
              <p className="truncate text-sm font-medium text-zinc-900">{displayName}</p>
              <p className="truncate text-xs text-zinc-500">LinuxDo ID: {linuxdoId || '未绑定'}</p>
              <p className="truncate text-xs text-zinc-500 capitalize">{user.role.replace('_', ' ')}</p>
            </div>
            <button
              onClick={logout}
              className="rounded-md p-1.5 text-zinc-400 hover:bg-white hover:text-red-600 hover:shadow-sm transition-all"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 pl-64">
        <div className="mx-auto max-w-7xl px-8 py-8">
          {children}
        </div>
      </main>
    </div>
  );
}

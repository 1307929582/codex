'use client';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { cn } from '@/lib/utils';
import { LayoutDashboard, Key, Activity, User, LogOut } from 'lucide-react';
import { useAuthStore } from '@/lib/stores/auth';
import { Button } from '@/components/ui/button';
import { useEffect } from 'react';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { logout, isAuthenticated, user } = useAuthStore();

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push('/login');
    }
  }, [isAuthenticated, router]);

  const navItems = [
    { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/keys', label: 'API Keys', icon: Key },
    { href: '/usage', label: 'Usage', icon: Activity },
    { href: '/account', label: 'Account', icon: User },
  ];

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (!isAuthenticated()) {
    return null;
  }

  return (
    <div className="flex min-h-screen">
      <aside className="w-64 border-r bg-gray-50/40 hidden md:block">
        <div className="flex h-16 items-center border-b px-6 font-bold text-lg">
          Codex Gateway
        </div>
        <nav className="p-4 space-y-2">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all hover:text-primary",
                  pathname === item.href ? "bg-gray-100 text-primary" : "text-muted-foreground"
                )}
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </Link>
            );
          })}
        </nav>
        <div className="absolute bottom-4 left-4 right-4">
          {user && (
            <div className="mb-2 px-3 py-2 text-sm text-muted-foreground truncate">
              {user.email}
            </div>
          )}
          <Button variant="outline" className="w-full justify-start" onClick={handleLogout}>
            <LogOut className="mr-2 h-4 w-4" /> Logout
          </Button>
        </div>
      </aside>
      <main className="flex-1 overflow-y-auto">
        <header className="flex h-16 items-center border-b px-6 md:hidden">
          <span className="font-bold">Codex Gateway</span>
        </header>
        <div className="p-8">
          {children}
        </div>
      </main>
    </div>
  );
}

'use client';

import { useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuthStore } from '@/lib/stores/auth';
import apiClient from '@/lib/api/client';

export default function AuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setAuth = useAuthStore((state) => state.setAuth);

  useEffect(() => {
    const token = searchParams.get('token');

    if (token) {
      // Fetch user info with the token
      apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;

      apiClient.get('/api/auth/me')
        .then((res) => {
          setAuth(token, res.data);
          router.push('/dashboard');
        })
        .catch((err) => {
          console.error('Failed to fetch user info:', err);
          router.push('/login?error=auth_failed');
        });
    } else {
      router.push('/login?error=no_token');
    }
  }, [searchParams, router, setAuth]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]" />
        <p className="mt-4 text-sm text-muted-foreground">正在登录...</p>
      </div>
    </div>
  );
}

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api/client';

export default function SetupRedirect({ children }: { children: React.ReactNode }) {
  const router = useRouter();

  useEffect(() => {
    const checkSetup = async () => {
      try {
        const response = await api.get('/api/setup/status');
        if (response.data.needs_setup) {
          // Redirect to setup wizard
          if (window.location.pathname !== '/setup') {
            router.push('/setup');
          }
        }
      } catch (error) {
        // If API fails, assume setup is needed
        if (window.location.pathname !== '/setup') {
          router.push('/setup');
        }
      }
    };

    checkSetup();
  }, [router]);

  return <>{children}</>;
}

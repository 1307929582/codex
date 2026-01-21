'use client';

import { useState } from 'react';
import apiClient from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function LoginPage() {
  const [error, setError] = useState('');

  const handleLinuxDoLogin = async () => {
    try {
      const res = await apiClient.get('/api/auth/linuxdo');
      window.location.href = res.data.url;
    } catch (err: any) {
      setError('LinuxDo 登录失败，请稍后重试');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-[400px]">
        <CardHeader>
          <CardTitle className="text-center text-2xl">Zenscale Codex</CardTitle>
          <p className="text-center text-sm text-muted-foreground">仅支持 LinuxDo 登录</p>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="text-sm text-red-500 text-center bg-red-50 p-2 rounded">{error}</div>
          )}
          <Button
            type="button"
            variant="outline"
            className="w-full"
            onClick={handleLinuxDoLogin}
          >
            <svg className="mr-2 h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z"/>
            </svg>
            使用 LinuxDo 登录
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

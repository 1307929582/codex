'use client';

import Link from 'next/link';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function RegisterPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-[400px]">
        <CardHeader>
          <CardTitle className="text-center text-2xl">注册已关闭</CardTitle>
          <p className="text-center text-sm text-muted-foreground">请使用 LinuxDo 登录</p>
        </CardHeader>
        <CardContent className="text-center">
          <Link href="/login" className="text-sm font-medium text-primary hover:underline">
            返回登录
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}

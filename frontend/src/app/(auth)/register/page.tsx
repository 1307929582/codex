'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useRouter } from 'next/navigation';
import { useState } from 'react';
import apiClient from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import Link from 'next/link';

const registerSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

type RegisterForm = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const router = useRouter();
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterForm) => {
    try {
      setError('');
      await apiClient.post('/api/auth/register', {
        email: data.email,
        password: data.password,
      });
      setSuccess(true);
      setTimeout(() => router.push('/login'), 2000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Registration failed');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-[400px]">
        <CardHeader>
          <CardTitle className="text-center text-2xl">Create Account</CardTitle>
          <p className="text-center text-sm text-muted-foreground">Sign up for Codex Gateway</p>
        </CardHeader>
        <CardContent>
          {success ? (
            <div className="text-center space-y-4">
              <div className="text-green-600 font-medium">Registration successful!</div>
              <p className="text-sm text-muted-foreground">Redirecting to login...</p>
            </div>
          ) : (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Email</label>
                <Input {...register('email')} placeholder="admin@example.com" type="email" />
                {errors.email && <p className="text-sm text-red-500">{errors.email.message}</p>}
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Password</label>
                <Input type="password" {...register('password')} placeholder="••••••••" />
                {errors.password && <p className="text-sm text-red-500">{errors.password.message}</p>}
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Confirm Password</label>
                <Input type="password" {...register('confirmPassword')} placeholder="••••••••" />
                {errors.confirmPassword && <p className="text-sm text-red-500">{errors.confirmPassword.message}</p>}
              </div>
              {error && <div className="text-sm text-red-500 text-center bg-red-50 p-2 rounded">{error}</div>}
              <Button type="submit" className="w-full" isLoading={isSubmitting}>
                Create Account
              </Button>
              <p className="text-center text-sm text-muted-foreground">
                Already have an account?{' '}
                <Link href="/login" className="text-primary hover:underline">
                  Sign in
                </Link>
              </p>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth';

export default function HomePage() {
  const router = useRouter();
  const { isAuthenticated, isLoading, context } = useAuth();

  useEffect(() => {
    if (isLoading) return;

    if (isAuthenticated) {
      if (context?.redirect_url) {
        router.replace(context.redirect_url);
      } else {
        router.replace('/org/new');
      }
    } else {
      router.replace('/login');
    }
  }, [isAuthenticated, isLoading, context, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="text-center">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-accent to-accent-hover mb-4">
          <svg
            className="w-8 h-8 text-white animate-pulse"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
            />
          </svg>
        </div>
        <h1 className="text-2xl font-bold text-foreground mb-2">Team Todo</h1>
        <div className="flex items-center justify-center gap-2 text-foreground-secondary">
          <div className="animate-spin rounded-full h-4 w-4 border-2 border-accent border-t-transparent" />
          読み込み中...
        </div>
      </div>
    </div>
  );
}

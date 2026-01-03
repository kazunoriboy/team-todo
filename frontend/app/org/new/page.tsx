'use client';

import { useState, FormEvent, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth';
import { organizationAPI, APIError } from '@/lib/api';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { slugify } from '@/lib/utils';

export default function NewOrganizationPage() {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [isSlugManual, setIsSlugManual] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.replace('/login');
    }
  }, [authLoading, isAuthenticated, router]);

  // Auto-generate slug from name
  useEffect(() => {
    if (!isSlugManual && name) {
      setSlug(slugify(name));
    }
  }, [name, isSlugManual]);

  const handleSlugChange = (value: string) => {
    setIsSlugManual(true);
    setSlug(slugify(value));
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name.trim()) {
      setError('組織名を入力してください。');
      return;
    }

    if (!slug.trim()) {
      setError('URLスラッグを入力してください。');
      return;
    }

    setIsLoading(true);

    try {
      const org = await organizationAPI.create(name.trim(), slug.trim());
      router.push(`/org/${org.slug}`);
    } catch (err) {
      if (err instanceof APIError) {
        if (err.message.includes('slug')) {
          setError('このURLスラッグはすでに使用されています。別のものを入力してください。');
        } else {
          setError(err.message);
        }
      } else {
        setError('組織の作成に失敗しました。もう一度お試しください。');
      }
    } finally {
      setIsLoading(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-8 w-8 border-2 border-accent border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      {/* Background decoration */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] bg-gradient-to-b from-accent/5 to-transparent rounded-full blur-3xl" />
      </div>

      <div className="w-full max-w-lg relative animate-fade-in">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-info to-info/70 mb-4">
            <svg
              className="w-8 h-8 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
              />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-foreground mb-2">組織を作成</h1>
          <p className="text-foreground-secondary">
            チームや会社の名前を入力してください
          </p>
        </div>

        {/* Form */}
        <div className="card">
          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="p-3 rounded-lg bg-error/10 border border-error/20 text-error text-sm animate-fade-in">
                {error}
              </div>
            )}

            <Input
              label="組織名"
              type="text"
              name="name"
              placeholder="株式会社〇〇、開発チームなど"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              autoFocus
            />

            <div>
              <Input
                label="URLスラッグ"
                type="text"
                name="slug"
                placeholder="my-team"
                value={slug}
                onChange={(e) => handleSlugChange(e.target.value)}
                required
              />
              <p className="mt-2 text-sm text-foreground-tertiary">
                組織のURLは <code className="px-1.5 py-0.5 rounded bg-background-tertiary text-foreground-secondary">teamtodo.com/org/{slug || 'your-slug'}</code> になります
              </p>
            </div>

            <div className="pt-2">
              <Button
                type="submit"
                className="w-full"
                size="lg"
                isLoading={isLoading}
              >
                組織を作成
              </Button>
            </div>
          </form>
        </div>

        {/* Info */}
        <div className="mt-6 p-4 rounded-xl bg-background-secondary border border-border">
          <div className="flex gap-3">
            <div className="flex-shrink-0">
              <svg
                className="w-5 h-5 text-info mt-0.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div>
              <h3 className="text-sm font-medium text-foreground mb-1">組織について</h3>
              <p className="text-sm text-foreground-secondary">
                組織を作成すると、自動的に「全般」プロジェクトが作成されます。
                後からチームメンバーを招待することができます。
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}


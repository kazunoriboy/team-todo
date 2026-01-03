'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { inviteAPI, organizationAPI, InviteInfo, APIError, getAccessToken } from '@/lib/api';
import { useAuth } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { formatDate } from '@/lib/utils';

export default function InvitePage() {
  const params = useParams();
  const router = useRouter();
  const token = params.token as string;
  const { isAuthenticated, isLoading: authLoading, refreshContext } = useAuth();
  
  const [inviteInfo, setInviteInfo] = useState<InviteInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAccepting, setIsAccepting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const loadInvite = async () => {
      try {
        const info = await inviteAPI.getInfo(token);
        setInviteInfo(info);
      } catch (err) {
        if (err instanceof APIError) {
          setError(err.message);
        } else {
          setError('招待情報の取得に失敗しました。');
        }
      } finally {
        setIsLoading(false);
      }
    };

    if (token) {
      loadInvite();
    }
  }, [token]);

  const handleAccept = async () => {
    setIsAccepting(true);
    setError('');

    try {
      const org = await organizationAPI.acceptInvite(token);
      await refreshContext();
      router.push(`/org/${org.slug}`);
    } catch (err) {
      if (err instanceof APIError) {
        setError(err.message);
      } else {
        setError('招待の承認に失敗しました。');
      }
    } finally {
      setIsAccepting(false);
    }
  };

  if (isLoading || authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-8 w-8 border-2 border-accent border-t-transparent" />
      </div>
    );
  }

  if (error && !inviteInfo) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <div className="w-full max-w-md text-center">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-error/10 flex items-center justify-center">
            <svg className="w-8 h-8 text-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-foreground mb-2">招待が無効です</h1>
          <p className="text-foreground-secondary mb-6">{error}</p>
          <Link href="/login">
            <Button>ログインページへ</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      {/* Background decoration */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-1/2 -left-1/2 w-full h-full bg-gradient-to-br from-success/5 to-transparent rounded-full blur-3xl" />
        <div className="absolute -bottom-1/2 -right-1/2 w-full h-full bg-gradient-to-tl from-accent/5 to-transparent rounded-full blur-3xl" />
      </div>

      <div className="w-full max-w-md relative animate-fade-in">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-success to-success/70 mb-4">
            <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-foreground mb-2">チームに参加</h1>
        </div>

        {/* Invite card */}
        <div className="card">
          {inviteInfo && (
            <>
              <div className="text-center pb-6 border-b border-border mb-6">
                <div className="w-16 h-16 mx-auto mb-4 rounded-xl bg-gradient-to-br from-info to-info/70 flex items-center justify-center text-white text-2xl font-bold">
                  {inviteInfo.organization_name.charAt(0).toUpperCase()}
                </div>
                <h2 className="text-xl font-semibold text-foreground mb-1">
                  {inviteInfo.organization_name}
                </h2>
                <p className="text-foreground-secondary text-sm">
                  への参加招待が届いています
                </p>
              </div>

              <div className="space-y-4 mb-6">
                <div className="flex items-center gap-3 text-sm">
                  <svg className="w-5 h-5 text-foreground-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  <span className="text-foreground-secondary">送信先:</span>
                  <span className="text-foreground">{inviteInfo.email}</span>
                </div>
                <div className="flex items-center gap-3 text-sm">
                  <svg className="w-5 h-5 text-foreground-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span className="text-foreground-secondary">有効期限:</span>
                  <span className="text-foreground">{formatDate(inviteInfo.expires_at)}</span>
                </div>
              </div>

              {error && (
                <div className="p-3 rounded-lg bg-error/10 border border-error/20 text-error text-sm mb-6">
                  {error}
                </div>
              )}

              {isAuthenticated ? (
                <Button
                  onClick={handleAccept}
                  className="w-full"
                  size="lg"
                  isLoading={isAccepting}
                >
                  招待を承認して参加
                </Button>
              ) : (
                <div className="space-y-3">
                  <p className="text-sm text-foreground-secondary text-center">
                    参加するにはログインまたは登録が必要です
                  </p>
                  <div className="flex gap-3">
                    <Link href={`/login?redirect=/invite/${token}`} className="flex-1">
                      <Button variant="secondary" className="w-full">
                        ログイン
                      </Button>
                    </Link>
                    <Link href={`/signup?redirect=/invite/${token}`} className="flex-1">
                      <Button className="w-full">
                        新規登録
                      </Button>
                    </Link>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}


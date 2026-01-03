'use client';

import { useState, FormEvent } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { projectAPI, APIError } from '@/lib/api';
import { cn } from '@/lib/utils';

interface CreateProjectModalProps {
  orgSlug: string;
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void;
}

export function CreateProjectModal({ orgSlug, isOpen, onClose, onCreated }: CreateProjectModalProps) {
  const [name, setName] = useState('');
  const [isPrivate, setIsPrivate] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name.trim()) {
      setError('プロジェクト名を入力してください。');
      return;
    }

    setIsLoading(true);

    try {
      await projectAPI.create(orgSlug, name.trim(), isPrivate);
      setName('');
      setIsPrivate(false);
      onCreated();
      onClose();
    } catch (err) {
      if (err instanceof APIError) {
        setError(err.message);
      } else {
        setError('プロジェクトの作成に失敗しました。');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    setName('');
    setIsPrivate(false);
    setError('');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-md bg-background-secondary border border-border rounded-xl shadow-modal animate-scale-in">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-lg font-semibold text-foreground">新しいプロジェクト</h2>
          <button
            onClick={handleClose}
            className="p-1.5 rounded-lg text-foreground-secondary hover:text-foreground hover:bg-background-hover transition-colors"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <form onSubmit={handleSubmit} className="p-4 space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-error/10 border border-error/20 text-error text-sm">
              {error}
            </div>
          )}

          <Input
            label="プロジェクト名"
            type="text"
            name="name"
            placeholder="例: マーケティングキャンペーン"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            autoFocus
          />

          {/* Privacy toggle */}
          <div>
            <label className="label">公開設定</label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setIsPrivate(false)}
                className={cn(
                  'flex-1 flex items-center gap-2 p-3 rounded-lg border transition-colors',
                  !isPrivate
                    ? 'border-accent bg-accent/10 text-foreground'
                    : 'border-border text-foreground-secondary hover:border-border-light'
                )}
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div className="text-left">
                  <p className="font-medium text-sm">公開</p>
                  <p className="text-xs text-foreground-tertiary">組織全体</p>
                </div>
              </button>
              <button
                type="button"
                onClick={() => setIsPrivate(true)}
                className={cn(
                  'flex-1 flex items-center gap-2 p-3 rounded-lg border transition-colors',
                  isPrivate
                    ? 'border-accent bg-accent/10 text-foreground'
                    : 'border-border text-foreground-secondary hover:border-border-light'
                )}
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <div className="text-left">
                  <p className="font-medium text-sm">非公開</p>
                  <p className="text-xs text-foreground-tertiary">招待のみ</p>
                </div>
              </button>
            </div>
          </div>
        </form>

        {/* Footer */}
        <div className="flex justify-end gap-3 p-4 border-t border-border">
          <Button type="button" variant="secondary" onClick={handleClose}>
            キャンセル
          </Button>
          <Button onClick={handleSubmit} isLoading={isLoading}>
            作成
          </Button>
        </div>
      </div>
    </div>
  );
}


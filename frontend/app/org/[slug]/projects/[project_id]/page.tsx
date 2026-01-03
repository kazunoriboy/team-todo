'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { projectAPI, Project } from '@/lib/api';
import { Button } from '@/components/ui/button';

export default function ProjectPage() {
  const params = useParams();
  const slug = params.slug as string;
  const projectId = params.project_id as string;
  
  const [project, setProject] = useState<Project | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadProject = async () => {
      try {
        const proj = await projectAPI.get(slug, projectId);
        setProject(proj);
      } catch (error) {
        console.error('Failed to load project:', error);
      } finally {
        setIsLoading(false);
      }
    };

    if (slug && projectId) {
      loadProject();
    }
  }, [slug, projectId]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-2 border-accent border-t-transparent" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-foreground mb-2">プロジェクトが見つかりません</h2>
          <p className="text-foreground-secondary">このプロジェクトは存在しないか、アクセス権限がありません。</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <header className="flex items-center justify-between px-8 py-4 border-b border-border bg-background-secondary">
        <div className="flex items-center gap-4">
          <div
            className={`w-10 h-10 rounded-lg flex items-center justify-center ${
              project.is_private
                ? 'bg-warning/10 text-warning'
                : 'bg-success/10 text-success'
            }`}
          >
            {project.is_private ? (
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            ) : (
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
            )}
          </div>
          <div>
            <h1 className="text-xl font-semibold text-foreground">{project.name}</h1>
            <p className="text-sm text-foreground-secondary">
              {project.is_private ? '非公開プロジェクト' : '公開プロジェクト'}
              {project.permission && ` · ${project.permission === 'edit' ? '編集可能' : '閲覧のみ'}`}
            </p>
          </div>
        </div>
        
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            メンバー
          </Button>
          <Button variant="secondary" size="sm">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            設定
          </Button>
        </div>
      </header>

      {/* Task views tabs */}
      <div className="flex items-center gap-1 px-8 py-2 border-b border-border bg-background">
        <button className="px-4 py-2 text-sm font-medium text-foreground border-b-2 border-accent">
          リスト
        </button>
        <button className="px-4 py-2 text-sm font-medium text-foreground-secondary hover:text-foreground transition-colors">
          ボード
        </button>
        <button className="px-4 py-2 text-sm font-medium text-foreground-secondary hover:text-foreground transition-colors">
          タイムライン
        </button>
        <button className="px-4 py-2 text-sm font-medium text-foreground-secondary hover:text-foreground transition-colors">
          カレンダー
        </button>
      </div>

      {/* Task list area */}
      <div className="flex-1 p-8 overflow-y-auto">
        {/* Add task button */}
        <button className="w-full flex items-center gap-3 px-4 py-3 rounded-lg border border-dashed border-border hover:border-border-light hover:bg-background-secondary transition-colors group mb-4">
          <div className="w-6 h-6 rounded-full border-2 border-border group-hover:border-accent group-hover:bg-accent flex items-center justify-center transition-colors">
            <svg className="w-3 h-3 text-foreground-tertiary group-hover:text-white transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <span className="text-foreground-secondary group-hover:text-foreground transition-colors">タスクを追加</span>
        </button>

        {/* Empty state */}
        <div className="text-center py-16">
          <div className="w-24 h-24 mx-auto mb-6 rounded-full bg-background-secondary flex items-center justify-center">
            <svg className="w-12 h-12 text-foreground-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-foreground mb-2">タスクがありません</h3>
          <p className="text-foreground-secondary mb-6 max-w-md mx-auto">
            このプロジェクトにはまだタスクがありません。<br />
            上のボタンをクリックして最初のタスクを作成しましょう。
          </p>
          <Button>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            タスクを作成
          </Button>
        </div>
      </div>
    </div>
  );
}


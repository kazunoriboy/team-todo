'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { useAuth } from '@/lib/auth';
import { Organization, Project } from '@/lib/api';

interface SidebarProps {
  organization: Organization;
  projects: Project[];
  onCreateProject: () => void;
}

export function Sidebar({ organization, projects, onCreateProject }: SidebarProps) {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);

  const isProjectActive = (projectId: string) => {
    return pathname.includes(`/projects/${projectId}`);
  };

  return (
    <aside className="w-64 h-screen bg-sidebar flex flex-col border-r border-border">
      {/* Organization Header */}
      <div className="p-4 border-b border-border">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-accent to-accent-hover flex items-center justify-center text-white font-semibold">
            {organization.name.charAt(0).toUpperCase()}
          </div>
          <div className="flex-1 min-w-0">
            <h2 className="font-semibold text-foreground truncate">{organization.name}</h2>
            <p className="text-xs text-foreground-tertiary">{organization.role}</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto p-3">
        {/* Main nav */}
        <div className="space-y-1 mb-6">
          <Link
            href={`/org/${organization.slug}`}
            className={cn(
              'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors',
              pathname === `/org/${organization.slug}`
                ? 'bg-sidebar-active text-foreground'
                : 'text-foreground-secondary hover:bg-sidebar-hover hover:text-foreground'
            )}
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
            </svg>
            ホーム
          </Link>
          <Link
            href={`/org/${organization.slug}/tasks`}
            className={cn(
              'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors',
              pathname.includes('/tasks')
                ? 'bg-sidebar-active text-foreground'
                : 'text-foreground-secondary hover:bg-sidebar-hover hover:text-foreground'
            )}
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
            マイタスク
          </Link>
        </div>

        {/* Projects section */}
        <div>
          <div className="flex items-center justify-between px-3 mb-2">
            <span className="text-xs font-medium text-foreground-tertiary uppercase tracking-wider">
              プロジェクト
            </span>
            <button
              onClick={onCreateProject}
              className="p-1 rounded hover:bg-sidebar-hover text-foreground-tertiary hover:text-foreground transition-colors"
              title="プロジェクトを作成"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>

          <div className="space-y-0.5">
            {projects.map((project) => (
              <Link
                key={project.id}
                href={`/org/${organization.slug}/projects/${project.id}`}
                className={cn(
                  'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors group',
                  isProjectActive(project.id)
                    ? 'bg-sidebar-active text-foreground'
                    : 'text-foreground-secondary hover:bg-sidebar-hover hover:text-foreground'
                )}
              >
                <span
                  className={cn(
                    'w-2 h-2 rounded-full',
                    project.is_private ? 'bg-warning' : 'bg-success'
                  )}
                />
                <span className="truncate flex-1">{project.name}</span>
                {project.is_private && (
                  <svg
                    className="w-3.5 h-3.5 text-foreground-tertiary opacity-0 group-hover:opacity-100 transition-opacity"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                )}
              </Link>
            ))}
          </div>
        </div>
      </nav>

      {/* User section */}
      <div className="p-3 border-t border-border relative">
        <button
          onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
          className="w-full flex items-center gap-3 p-2 rounded-lg hover:bg-sidebar-hover transition-colors"
        >
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-info to-info/70 flex items-center justify-center text-white text-sm font-medium">
            {user?.display_name?.charAt(0).toUpperCase() || 'U'}
          </div>
          <div className="flex-1 text-left min-w-0">
            <p className="text-sm font-medium text-foreground truncate">{user?.display_name}</p>
            <p className="text-xs text-foreground-tertiary truncate">{user?.email}</p>
          </div>
          <svg
            className={cn(
              'w-4 h-4 text-foreground-tertiary transition-transform',
              isUserMenuOpen && 'rotate-180'
            )}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        {/* User dropdown menu */}
        {isUserMenuOpen && (
          <div className="absolute bottom-full left-3 right-3 mb-2 bg-background-secondary border border-border rounded-lg shadow-modal overflow-hidden animate-scale-in">
            <Link
              href="/settings"
              className="flex items-center gap-3 px-4 py-3 text-sm text-foreground-secondary hover:bg-background-hover hover:text-foreground transition-colors"
              onClick={() => setIsUserMenuOpen(false)}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              設定
            </Link>
            <Link
              href="/org/new"
              className="flex items-center gap-3 px-4 py-3 text-sm text-foreground-secondary hover:bg-background-hover hover:text-foreground transition-colors"
              onClick={() => setIsUserMenuOpen(false)}
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              新しい組織を作成
            </Link>
            <div className="border-t border-border">
              <button
                onClick={() => {
                  setIsUserMenuOpen(false);
                  logout();
                }}
                className="w-full flex items-center gap-3 px-4 py-3 text-sm text-error hover:bg-error/10 transition-colors"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                </svg>
                ログアウト
              </button>
            </div>
          </div>
        )}
      </div>
    </aside>
  );
}


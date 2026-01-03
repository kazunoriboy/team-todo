'use client';

import { useEffect, useState, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth';
import { organizationAPI, projectAPI, Organization, Project } from '@/lib/api';
import { Sidebar } from '@/components/layout/sidebar';
import { CreateProjectModal } from '@/components/layout/create-project-modal';

export default function OrgLayout({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const router = useRouter();
  const slug = params.slug as string;
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateProjectOpen, setIsCreateProjectOpen] = useState(false);

  const loadData = useCallback(async () => {
    try {
      const [org, projectList] = await Promise.all([
        organizationAPI.get(slug),
        projectAPI.list(slug),
      ]);
      setOrganization(org);
      setProjects(projectList);
    } catch (error) {
      console.error('Failed to load organization:', error);
      router.push('/org/new');
    } finally {
      setIsLoading(false);
    }
  }, [slug, router]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.replace('/login');
      return;
    }

    if (isAuthenticated && slug) {
      loadData();
    }
  }, [authLoading, isAuthenticated, slug, router, loadData]);

  const handleProjectCreated = () => {
    loadData();
  };

  if (authLoading || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-8 w-8 border-2 border-accent border-t-transparent" />
      </div>
    );
  }

  if (!organization) {
    return null;
  }

  return (
    <div className="flex h-screen bg-background overflow-hidden">
      <Sidebar
        organization={organization}
        projects={projects}
        onCreateProject={() => setIsCreateProjectOpen(true)}
      />
      
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>

      <CreateProjectModal
        orgSlug={slug}
        isOpen={isCreateProjectOpen}
        onClose={() => setIsCreateProjectOpen(false)}
        onCreated={handleProjectCreated}
      />
    </div>
  );
}


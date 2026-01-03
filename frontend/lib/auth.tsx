'use client';

import { createContext, useContext, useEffect, useState, useCallback, ReactNode } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { authAPI, contextAPI, getAccessToken, clearTokens, User, ContextResponse } from './api';

interface AuthContextType {
  user: User | null;
  context: ContextResponse | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => void;
  refreshContext: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Routes that don't require authentication
const publicRoutes = ['/login', '/signup', '/invite'];

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [context, setContext] = useState<ContextResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();

  const isAuthenticated = !!user;

  const refreshContext = useCallback(async () => {
    try {
      const ctx = await contextAPI.getCurrent();
      setContext(ctx);
      return ctx;
    } catch (error) {
      console.error('Failed to refresh context:', error);
      return null;
    }
  }, []);

  // Check authentication on mount
  useEffect(() => {
    const checkAuth = async () => {
      const token = getAccessToken();
      
      if (!token) {
        setIsLoading(false);
        // Redirect to login if trying to access protected route
        if (!publicRoutes.some(route => pathname.startsWith(route)) && pathname !== '/') {
          router.replace('/login');
        }
        return;
      }

      try {
        const userData = await authAPI.getMe();
        setUser(userData);
        
        // Get user context
        const ctx = await contextAPI.getCurrent();
        setContext(ctx);
        
        // Redirect from root or login pages if authenticated
        if (pathname === '/' || pathname === '/login' || pathname === '/signup') {
          if (ctx.redirect_url) {
            router.replace(ctx.redirect_url);
          }
        }
      } catch (error) {
        console.error('Auth check failed:', error);
        clearTokens();
        if (!publicRoutes.some(route => pathname.startsWith(route)) && pathname !== '/') {
          router.replace('/login');
        }
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, [pathname, router, refreshContext]);

  const login = async (email: string, password: string) => {
    const response = await authAPI.login(email, password);
    setUser(response.user);
    
    // Get context and redirect
    const ctx = await contextAPI.getCurrent();
    setContext(ctx);
    
    if (ctx.redirect_url) {
      router.push(ctx.redirect_url);
    } else {
      router.push('/org/new');
    }
  };

  const register = async (email: string, password: string, displayName: string) => {
    const response = await authAPI.register(email, password, displayName);
    setUser(response.user);
    setContext(null);
    // New users need to create an organization
    router.push('/org/new');
  };

  const logout = () => {
    authAPI.logout();
    setUser(null);
    setContext(null);
    router.push('/login');
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        context,
        isLoading,
        isAuthenticated,
        login,
        register,
        logout,
        refreshContext,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// HOC for protected pages
export function withAuth<P extends object>(
  WrappedComponent: React.ComponentType<P>
) {
  return function ProtectedComponent(props: P) {
    const { isAuthenticated, isLoading } = useAuth();
    const router = useRouter();

    useEffect(() => {
      if (!isLoading && !isAuthenticated) {
        router.replace('/login');
      }
    }, [isLoading, isAuthenticated, router]);

    if (isLoading) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-background">
          <div className="animate-spin rounded-full h-8 w-8 border-2 border-accent border-t-transparent" />
        </div>
      );
    }

    if (!isAuthenticated) {
      return null;
    }

    return <WrappedComponent {...props} />;
  };
}


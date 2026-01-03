const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Types
export interface User {
  id: string;
  email: string;
  display_name: string;
  last_org_id?: string;
  last_project_id?: string;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  role?: string;
  created_at: string;
}

export interface Project {
  id: string;
  name: string;
  is_private: boolean;
  organization_id: string;
  permission?: string;
  created_at: string;
}

export interface ContextResponse {
  has_context: boolean;
  organization?: Organization;
  project?: Project;
  redirect_url?: string;
}

export interface InviteInfo {
  organization_name: string;
  organization_slug: string;
  email: string;
  expires_at: string;
}

// API Error
export class APIError extends Error {
  status: number;
  
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = 'APIError';
  }
}

// Token management
const TOKEN_KEY = 'team_todo_access_token';
const REFRESH_TOKEN_KEY = 'team_todo_refresh_token';

export const getAccessToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
};

export const getRefreshToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(REFRESH_TOKEN_KEY);
};

export const setTokens = (accessToken: string, refreshToken: string) => {
  localStorage.setItem(TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
};

export const clearTokens = () => {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
};

// API client
async function fetchWithAuth<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const accessToken = getAccessToken();
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };
  
  if (accessToken) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${accessToken}`;
  }
  
  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers,
  });
  
  if (!response.ok) {
    // Try to get error message from response
    let message = 'An error occurred';
    try {
      const data = await response.json();
      message = data.message || data.error || message;
    } catch {
      // Ignore JSON parse errors
    }
    
    // Handle token expiration
    if (response.status === 401 && accessToken) {
      const refreshed = await refreshAccessToken();
      if (refreshed) {
        // Retry the request with new token
        return fetchWithAuth<T>(endpoint, options);
      }
      // Clear tokens and redirect to login
      clearTokens();
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }
    
    throw new APIError(message, response.status);
  }
  
  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }
  
  return response.json();
}

async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;
  
  try {
    const response = await fetch(`${API_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    
    if (!response.ok) return false;
    
    const data: AuthResponse = await response.json();
    setTokens(data.access_token, data.refresh_token);
    return true;
  } catch {
    return false;
  }
}

// Auth API
export const authAPI = {
  register: async (email: string, password: string, displayName: string): Promise<AuthResponse> => {
    const response = await fetch(`${API_URL}/api/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password, display_name: displayName }),
    });
    
    if (!response.ok) {
      const data = await response.json();
      throw new APIError(data.message || 'Registration failed', response.status);
    }
    
    const data: AuthResponse = await response.json();
    setTokens(data.access_token, data.refresh_token);
    return data;
  },
  
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const response = await fetch(`${API_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    
    if (!response.ok) {
      const data = await response.json();
      throw new APIError(data.message || 'Login failed', response.status);
    }
    
    const data: AuthResponse = await response.json();
    setTokens(data.access_token, data.refresh_token);
    return data;
  },
  
  logout: () => {
    clearTokens();
  },
  
  getMe: (): Promise<User> => fetchWithAuth('/api/v1/me'),
  
  updateMe: (data: { display_name?: string }): Promise<User> => 
    fetchWithAuth('/api/v1/me', { method: 'PATCH', body: JSON.stringify(data) }),
};

// Context API
export const contextAPI = {
  getCurrent: (): Promise<ContextResponse> => fetchWithAuth('/api/v1/context'),
  
  update: (data: { org_id?: string; project_id?: string }): Promise<void> =>
    fetchWithAuth('/api/v1/context', { method: 'PUT', body: JSON.stringify(data) }),
};

// Organization API
export const organizationAPI = {
  create: (name: string, slug: string): Promise<Organization> =>
    fetchWithAuth('/api/v1/organizations', {
      method: 'POST',
      body: JSON.stringify({ name, slug }),
    }),
  
  list: (): Promise<Organization[]> => fetchWithAuth('/api/v1/organizations'),
  
  get: (slug: string): Promise<Organization> =>
    fetchWithAuth(`/api/v1/organizations/${slug}`),
  
  invite: (slug: string, email: string, role: string): Promise<{ id: string; email: string; role: string }> =>
    fetchWithAuth(`/api/v1/organizations/${slug}/invites`, {
      method: 'POST',
      body: JSON.stringify({ email, role }),
    }),
  
  acceptInvite: (token: string): Promise<Organization> =>
    fetchWithAuth(`/api/v1/invites/${token}/accept`, { method: 'POST' }),
};

// Invite API (public)
export const inviteAPI = {
  getInfo: async (token: string): Promise<InviteInfo> => {
    const response = await fetch(`${API_URL}/api/v1/invites/${token}`);
    if (!response.ok) {
      const data = await response.json();
      throw new APIError(data.message || 'Failed to get invite info', response.status);
    }
    return response.json();
  },
};

// Project API
export const projectAPI = {
  create: (orgSlug: string, name: string, isPrivate: boolean): Promise<Project> =>
    fetchWithAuth(`/api/v1/organizations/${orgSlug}/projects`, {
      method: 'POST',
      body: JSON.stringify({ name, is_private: isPrivate }),
    }),
  
  list: (orgSlug: string): Promise<Project[]> =>
    fetchWithAuth(`/api/v1/organizations/${orgSlug}/projects`),
  
  get: (orgSlug: string, projectId: string): Promise<Project> =>
    fetchWithAuth(`/api/v1/organizations/${orgSlug}/projects/${projectId}`),
  
  addMember: (orgSlug: string, projectId: string, userId: string, permission: string): Promise<void> =>
    fetchWithAuth(`/api/v1/organizations/${orgSlug}/projects/${projectId}/members`, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, permission }),
    }),
};


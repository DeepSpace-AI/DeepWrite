export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  institution?: string;
  research_fields?: string[];
  subscription_tier: string;
  created_at: string;
}

export interface Team {
  id: string;
  name: string;
  description?: string;
  owner_id: string;
  max_members: number;
  created_at: string;
}

export interface Project {
  id: string;
  title: string;
  description?: string;
  owner_id: string;
  status: string;
  research_field?: string;
  is_archived: boolean;
  created_at: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  meta?: {
    page?: number;
    limit?: number;
    total?: number;
  };
}

export interface ApiError {
  success: false;
  error: {
    code: string;
    message: string;
  };
}

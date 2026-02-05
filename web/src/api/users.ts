import apiClient from './client';
import type { User, Tenant, PaginatedResponse } from '@/types';

export interface UserListParams {
  page?: number;
  page_size?: number;
}

export interface UpdateUserRequest {
  first_name?: string;
  last_name?: string;
  status?: string;
}

export interface CreateTenantRequest {
  name: string;
  slug: string;
  contact_email: string;
  subscription_tier?: 'free' | 'basic' | 'professional' | 'enterprise';
  max_devices?: number;
  max_users?: number;
}

export interface UpdateTenantRequest {
  name?: string;
  contact_email?: string;
  subscription_tier?: string;
  max_devices?: number;
  max_users?: number;
  status?: string;
}

export const usersApi = {
  list: async (params?: UserListParams): Promise<PaginatedResponse<User>> => {
    const response = await apiClient.get<PaginatedResponse<User>>('/users', { params });
    return response.data;
  },

  get: async (id: string): Promise<User> => {
    const response = await apiClient.get<User>(`/users/${id}`);
    return response.data;
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await apiClient.get<User>('/users/me');
    return response.data;
  },

  update: async (id: string, data: UpdateUserRequest): Promise<User> => {
    const response = await apiClient.put<User>(`/users/${id}`, data);
    return response.data;
  },
};

export const tenantsApi = {
  list: async (params?: UserListParams): Promise<PaginatedResponse<Tenant>> => {
    const response = await apiClient.get<PaginatedResponse<Tenant>>('/tenants', { params });
    return response.data;
  },

  get: async (id: string): Promise<Tenant> => {
    const response = await apiClient.get<Tenant>(`/tenants/${id}`);
    return response.data;
  },

  create: async (data: CreateTenantRequest): Promise<Tenant> => {
    const response = await apiClient.post<Tenant>('/tenants', data);
    return response.data;
  },

  update: async (id: string, data: UpdateTenantRequest): Promise<Tenant> => {
    const response = await apiClient.put<Tenant>(`/tenants/${id}`, data);
    return response.data;
  },
};

export default { usersApi, tenantsApi };

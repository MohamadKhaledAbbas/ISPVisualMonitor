import apiClient from './client';
import type {
  Router,
  CreateRouterRequest,
  UpdateRouterRequest,
  PaginatedResponse,
  NetworkInterface,
} from '@/types';

export interface RouterListParams {
  page?: number;
  page_size?: number;
  status?: string;
  search?: string;
}

export const routersApi = {
  list: async (params?: RouterListParams): Promise<PaginatedResponse<Router>> => {
    const response = await apiClient.get<PaginatedResponse<Router>>('/routers', { params });
    return response.data;
  },

  get: async (id: string): Promise<Router> => {
    const response = await apiClient.get<Router>(`/routers/${id}`);
    return response.data;
  },

  create: async (data: CreateRouterRequest): Promise<Router> => {
    const response = await apiClient.post<Router>('/routers', data);
    return response.data;
  },

  update: async (id: string, data: UpdateRouterRequest): Promise<Router> => {
    const response = await apiClient.put<Router>(`/routers/${id}`, data);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/routers/${id}`);
  },

  bulkUpdatePolling: async (ids: string[], enabled: boolean): Promise<void> => {
    await apiClient.post('/routers/bulk/polling', { router_ids: ids, polling_enabled: enabled });
  },

  bulkDelete: async (ids: string[]): Promise<void> => {
    await apiClient.post('/routers/bulk/delete', { router_ids: ids });
  },

  getInterfaces: async (routerId: string, params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<NetworkInterface>> => {
    const response = await apiClient.get<PaginatedResponse<NetworkInterface>>(
      `/routers/${routerId}/interfaces`,
      { params }
    );
    return response.data;
  },
};

export default routersApi;

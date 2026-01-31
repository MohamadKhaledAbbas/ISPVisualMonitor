import apiClient from './client';
import type { InterfaceMetrics, RouterMetrics, DashboardStats, MetricDataPoint } from '@/types';

export interface MetricsQueryParams {
  from?: string;
  to?: string;
}

export const metricsApi = {
  getInterfaceMetrics: async (interfaceId: string, params?: MetricsQueryParams): Promise<InterfaceMetrics> => {
    const response = await apiClient.get<InterfaceMetrics>(`/metrics/interfaces/${interfaceId}`, { params });
    return response.data;
  },

  getRouterMetrics: async (routerId: string, params?: MetricsQueryParams): Promise<RouterMetrics> => {
    const response = await apiClient.get<RouterMetrics>(`/metrics/routers/${routerId}`, { params });
    return response.data;
  },

  getDashboardStats: async (): Promise<DashboardStats> => {
    const response = await apiClient.get<DashboardStats>('/metrics/dashboard');
    return response.data;
  },

  getNetworkTraffic: async (params?: MetricsQueryParams): Promise<{ in_bps: MetricDataPoint[]; out_bps: MetricDataPoint[] }> => {
    const response = await apiClient.get('/metrics/traffic', { params });
    return response.data;
  },

  exportMetrics: async (routerId: string, params?: MetricsQueryParams & { format?: 'csv' | 'json' }): Promise<Blob> => {
    const response = await apiClient.get(`/metrics/routers/${routerId}/export`, {
      params,
      responseType: 'blob',
    });
    return response.data;
  },
};

export default metricsApi;

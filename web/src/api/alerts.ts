import apiClient from './client';
import type {
  Alert,
  AlertSeverity,
  AlertStatus,
  AcknowledgeAlertRequest,
  PaginatedResponse,
} from '@/types';

export interface AlertListParams {
  page?: number;
  page_size?: number;
  severity?: AlertSeverity;
  status?: AlertStatus;
  router_id?: string;
}

export const alertsApi = {
  list: async (params?: AlertListParams): Promise<PaginatedResponse<Alert>> => {
    const response = await apiClient.get<PaginatedResponse<Alert>>('/alerts', { params });
    return response.data;
  },

  get: async (id: string): Promise<Alert> => {
    const response = await apiClient.get<Alert>(`/alerts/${id}`);
    return response.data;
  },

  acknowledge: async (id: string, data?: AcknowledgeAlertRequest): Promise<void> => {
    await apiClient.post(`/alerts/${id}/acknowledge`, data || {});
  },

  resolve: async (id: string): Promise<void> => {
    await apiClient.post(`/alerts/${id}/resolve`);
  },

  getStatistics: async (): Promise<{ total: number; by_severity: Record<AlertSeverity, number>; by_status: Record<AlertStatus, number> }> => {
    const response = await apiClient.get('/alerts/statistics');
    return response.data;
  },
};

export default alertsApi;

import apiClient from './client';
import type { PPPoESession, DHCPLease, NATConnection, PaginatedResponse } from '@/types';

export interface SessionListParams {
  page?: number;
  page_size?: number;
  router_id?: string;
  search?: string;
}

export const sessionsApi = {
  listPPPoE: async (params?: SessionListParams): Promise<PaginatedResponse<PPPoESession>> => {
    const response = await apiClient.get<PaginatedResponse<PPPoESession>>('/sessions/pppoe', { params });
    return response.data;
  },

  listDHCP: async (params?: SessionListParams): Promise<PaginatedResponse<DHCPLease>> => {
    const response = await apiClient.get<PaginatedResponse<DHCPLease>>('/sessions/dhcp', { params });
    return response.data;
  },

  listNAT: async (params?: SessionListParams): Promise<PaginatedResponse<NATConnection>> => {
    const response = await apiClient.get<PaginatedResponse<NATConnection>>('/sessions/nat', { params });
    return response.data;
  },

  terminatePPPoE: async (sessionId: string): Promise<void> => {
    await apiClient.post(`/sessions/pppoe/${sessionId}/terminate`);
  },
};

export default sessionsApi;

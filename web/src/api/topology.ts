import apiClient from './client';
import type { Topology, TopologyGeoJSON } from '@/types';

export const topologyApi = {
  getTopology: async (): Promise<Topology> => {
    const response = await apiClient.get<Topology>('/topology');
    return response.data;
  },

  getGeoJSON: async (): Promise<TopologyGeoJSON> => {
    const response = await apiClient.get<TopologyGeoJSON>('/topology/geojson');
    return response.data;
  },
};

export default topologyApi;

// User and Auth types
export interface User {
  id: string;
  tenant_id: string;
  email: string;
  first_name: string;
  last_name: string;
  status: 'active' | 'inactive' | 'pending';
  email_verified: boolean;
  roles?: string[];
  created_at: string;
  updated_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface LoginResponse extends AuthTokens {
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  tenant_id?: string;
}

// Tenant types
export interface Tenant {
  id: string;
  name: string;
  slug: string;
  contact_email: string;
  subscription_tier: 'free' | 'basic' | 'professional' | 'enterprise';
  max_devices: number;
  max_users: number;
  status: 'active' | 'inactive' | 'suspended';
  created_at: string;
  updated_at: string;
}

// Router types
export interface Location {
  latitude: number;
  longitude: number;
}

export interface Router {
  id: string;
  tenant_id: string;
  name: string;
  hostname: string;
  management_ip: string;
  vendor: string;
  model: string;
  status: 'active' | 'inactive' | 'maintenance' | 'offline';
  location?: Location;
  polling_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateRouterRequest {
  name: string;
  hostname: string;
  management_ip: string;
  vendor?: string;
  model?: string;
  location?: Location;
}

export interface UpdateRouterRequest {
  name?: string;
  hostname?: string;
  management_ip?: string;
  vendor?: string;
  model?: string;
  status?: Router['status'];
  location?: Location;
  polling_enabled?: boolean;
}

// Interface types
export interface NetworkInterface {
  id: string;
  router_id: string;
  name: string;
  description?: string;
  if_index: number;
  if_type: string;
  admin_status: 'up' | 'down';
  oper_status: 'up' | 'down';
  speed: number;
  mtu: number;
  mac_address?: string;
  ip_addresses?: string[];
  created_at: string;
  updated_at: string;
}

// Topology types
export interface TopologyLink {
  id: string;
  source_interface_id: string;
  target_interface_id: string;
  source_router_id: string;
  target_router_id: string;
  link_type: 'ethernet' | 'fiber' | 'wireless' | 'vpn';
  status: 'active' | 'inactive' | 'degraded';
}

export interface Topology {
  routers: Router[];
  links: TopologyLink[];
}

export interface TopologyGeoJSON {
  type: 'FeatureCollection';
  features: GeoJSONFeature[];
}

export interface GeoJSONFeature {
  type: 'Feature';
  geometry: {
    type: 'Point' | 'LineString';
    coordinates: number[] | number[][];
  };
  properties: {
    type: 'router' | 'link';
    id: string;
    name?: string;
    status?: string;
    [key: string]: unknown;
  };
}

// Metrics types
export interface MetricDataPoint {
  timestamp: string;
  value: number;
}

export interface InterfaceMetrics {
  interface_id: string;
  interface_name: string;
  in_bps: MetricDataPoint[];
  out_bps: MetricDataPoint[];
  utilization: MetricDataPoint[];
  in_errors: MetricDataPoint[];
  out_errors: MetricDataPoint[];
}

export interface RouterMetrics {
  router_id: string;
  router_name: string;
  cpu_usage: MetricDataPoint[];
  memory_usage: MetricDataPoint[];
  temperature: MetricDataPoint[];
  uptime: number;
}

// Alert types
export type AlertSeverity = 'critical' | 'warning' | 'info';
export type AlertStatus = 'open' | 'acknowledged' | 'resolved';

export interface Alert {
  id: string;
  tenant_id: string;
  router_id?: string;
  interface_id?: string;
  severity: AlertSeverity;
  status: AlertStatus;
  title: string;
  description: string;
  acknowledged_by?: string;
  acknowledged_at?: string;
  acknowledged_note?: string;
  resolved_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AcknowledgeAlertRequest {
  note?: string;
}

// Session types
export interface PPPoESession {
  id: string;
  router_id: string;
  username: string;
  ip_address: string;
  mac_address: string;
  interface_name: string;
  uptime: number;
  rx_bytes: number;
  tx_bytes: number;
  session_id: string;
  created_at: string;
}

export interface DHCPLease {
  id: string;
  router_id: string;
  ip_address: string;
  mac_address: string;
  hostname?: string;
  lease_time: number;
  expires_at: string;
  created_at: string;
}

export interface NATConnection {
  id: string;
  router_id: string;
  protocol: 'tcp' | 'udp' | 'icmp';
  src_address: string;
  dst_address: string;
  src_port?: number;
  dst_port?: number;
  reply_src_address?: string;
  reply_dst_address?: string;
  reply_src_port?: number;
  reply_dst_port?: number;
  state: string;
  timeout: number;
  created_at: string;
}

// Pagination types
export interface Pagination {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: Pagination;
}

// API Error types
export interface APIError {
  code: string;
  message: string;
  details?: string | ValidationError[];
}

export interface ValidationError {
  field: string;
  message: string;
}

// Dashboard stats types
export interface DashboardStats {
  total_routers: number;
  online_routers: number;
  offline_routers: number;
  active_alerts: number;
  active_sessions: number;
  total_bandwidth: {
    in_bps: number;
    out_bps: number;
  };
}

// Time range options
export type TimeRange = '1h' | '6h' | '24h' | '7d' | '30d' | 'custom';

export interface CustomTimeRange {
  from: string;
  to: string;
}

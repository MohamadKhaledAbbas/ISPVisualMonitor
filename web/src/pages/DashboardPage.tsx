import { useQuery } from '@tanstack/react-query';
import {
  ServerIcon,
  SignalIcon,
  ExclamationTriangleIcon,
  ArrowTrendingUpIcon,
  ArrowTrendingDownIcon,
} from '@heroicons/react/24/outline';
import { Card, StatCard, Badge } from '@/components/common';
import { metricsApi } from '@/api';
import { formatBps, formatDuration } from '@/utils';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { Link } from 'react-router-dom';
import { format } from 'date-fns';
import type { Alert, Router } from '@/types';

// Mock data for development
const mockStats = {
  total_routers: 24,
  online_routers: 21,
  offline_routers: 3,
  active_alerts: 5,
  active_sessions: 1250,
  total_bandwidth: {
    in_bps: 4500000000,
    out_bps: 2300000000,
  },
};

const mockTrafficData = Array.from({ length: 24 }, (_, i) => ({
  time: format(new Date(Date.now() - (23 - i) * 3600000), 'HH:mm'),
  in_bps: Math.random() * 5000000000 + 2000000000,
  out_bps: Math.random() * 2500000000 + 1000000000,
}));

const mockRouterStatusData = [
  { name: 'Online', value: 21, color: '#10B981' },
  { name: 'Offline', value: 3, color: '#EF4444' },
];

const mockRecentAlerts: Alert[] = [
  {
    id: '1',
    tenant_id: '1',
    router_id: 'r1',
    severity: 'critical',
    status: 'open',
    title: 'Router R1 offline',
    description: 'Router is not responding to SNMP polls',
    created_at: new Date(Date.now() - 300000).toISOString(),
    updated_at: new Date(Date.now() - 300000).toISOString(),
  },
  {
    id: '2',
    tenant_id: '1',
    router_id: 'r2',
    severity: 'warning',
    status: 'open',
    title: 'High CPU usage on R2',
    description: 'CPU utilization above 90%',
    created_at: new Date(Date.now() - 600000).toISOString(),
    updated_at: new Date(Date.now() - 600000).toISOString(),
  },
  {
    id: '3',
    tenant_id: '1',
    interface_id: 'if1',
    severity: 'info',
    status: 'acknowledged',
    title: 'Interface flapping on R3',
    description: 'GigabitEthernet0/1 status changed multiple times',
    created_at: new Date(Date.now() - 1800000).toISOString(),
    updated_at: new Date(Date.now() - 1800000).toISOString(),
  },
];

const mockTopRouters: Router[] = [
  {
    id: '1',
    tenant_id: '1',
    name: 'Core-Router-1',
    hostname: 'core1.example.com',
    management_ip: '10.0.0.1',
    vendor: 'Cisco',
    model: 'ASR 9000',
    status: 'active',
    polling_enabled: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '2',
    tenant_id: '1',
    name: 'Edge-Router-1',
    hostname: 'edge1.example.com',
    management_ip: '10.0.0.2',
    vendor: 'MikroTik',
    model: 'CCR1036',
    status: 'active',
    polling_enabled: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

export function DashboardPage() {
  // In production, fetch from API
  const { data: stats } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: () => metricsApi.getDashboardStats(),
    enabled: false, // Using mock data for now
  });

  const dashboardStats = stats || mockStats;

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Network overview and key metrics
        </p>
      </div>

      {/* Stats cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Routers"
          value={dashboardStats.total_routers}
          icon={<ServerIcon className="h-6 w-6 text-blue-600" />}
        />
        <StatCard
          title="Online"
          value={dashboardStats.online_routers}
          icon={<SignalIcon className="h-6 w-6 text-green-600" />}
          change={{ value: 2, type: 'increase' }}
        />
        <StatCard
          title="Offline"
          value={dashboardStats.offline_routers}
          icon={<ExclamationTriangleIcon className="h-6 w-6 text-red-600" />}
        />
        <StatCard
          title="Active Alerts"
          value={dashboardStats.active_alerts}
          icon={<ExclamationTriangleIcon className="h-6 w-6 text-yellow-600" />}
        />
      </div>

      {/* Bandwidth stats */}
      <div className="grid gap-4 sm:grid-cols-2">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Total Inbound</p>
              <p className="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
                {formatBps(dashboardStats.total_bandwidth.in_bps)}
              </p>
            </div>
            <ArrowTrendingDownIcon className="h-8 w-8 text-green-500" />
          </div>
        </Card>
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Total Outbound</p>
              <p className="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
                {formatBps(dashboardStats.total_bandwidth.out_bps)}
              </p>
            </div>
            <ArrowTrendingUpIcon className="h-8 w-8 text-blue-500" />
          </div>
        </Card>
      </div>

      {/* Charts section */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Traffic chart */}
        <Card className="lg:col-span-2" title="Network Traffic (24h)">
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={mockTrafficData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
                <XAxis 
                  dataKey="time" 
                  className="text-xs text-gray-500"
                  tick={{ fill: '#6B7280' }}
                />
                <YAxis 
                  tickFormatter={(value) => formatBps(value)}
                  className="text-xs text-gray-500"
                  tick={{ fill: '#6B7280' }}
                />
                <Tooltip 
                  formatter={(value) => typeof value === 'number' ? formatBps(value) : value}
                  contentStyle={{
                    backgroundColor: 'var(--tooltip-bg, #fff)',
                    borderColor: 'var(--tooltip-border, #e5e7eb)',
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="in_bps"
                  stroke="#10B981"
                  strokeWidth={2}
                  dot={false}
                  name="Inbound"
                />
                <Line
                  type="monotone"
                  dataKey="out_bps"
                  stroke="#3B82F6"
                  strokeWidth={2}
                  dot={false}
                  name="Outbound"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Router status pie chart */}
        <Card title="Router Status">
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={mockRouterStatusData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {mockRouterStatusData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
            <div className="flex justify-center gap-6 mt-4">
              {mockRouterStatusData.map((entry) => (
                <div key={entry.name} className="flex items-center gap-2">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: entry.color }}
                  />
                  <span className="text-sm text-gray-600 dark:text-gray-400">
                    {entry.name}: {entry.value}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </div>

      {/* Recent alerts and top routers */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Recent alerts */}
        <Card
          title="Recent Alerts"
          actions={
            <Link
              to="/alerts"
              className="text-sm font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400"
            >
              View all
            </Link>
          }
        >
          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {mockRecentAlerts.map((alert) => (
              <div key={alert.id} className="py-3 flex items-start justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white">
                    {alert.title}
                  </p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {alert.description}
                  </p>
                  <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                    {formatDuration((Date.now() - new Date(alert.created_at).getTime()) / 1000)} ago
                  </p>
                </div>
                <Badge
                  variant={
                    alert.severity === 'critical'
                      ? 'danger'
                      : alert.severity === 'warning'
                      ? 'warning'
                      : 'info'
                  }
                  size="sm"
                >
                  {alert.severity}
                </Badge>
              </div>
            ))}
          </div>
        </Card>

        {/* Top routers */}
        <Card
          title="Top Routers by Traffic"
          actions={
            <Link
              to="/routers"
              className="text-sm font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400"
            >
              View all
            </Link>
          }
        >
          <div className="divide-y divide-gray-200 dark:divide-gray-700">
            {mockTopRouters.map((router) => (
              <div key={router.id} className="py-3 flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white">
                    {router.name}
                  </p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {router.management_ip}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium text-gray-900 dark:text-white">
                    {formatBps(Math.random() * 5000000000)}
                  </p>
                  <Badge
                    variant={router.status === 'active' ? 'success' : 'danger'}
                    size="sm"
                  >
                    {router.status}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </div>
  );
}

export default DashboardPage;

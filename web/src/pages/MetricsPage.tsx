import { useState } from 'react';
import { Card, Button, Select } from '@/components/common';
import type { TimeRange } from '@/types';
import { formatBps, formatPercentage, downloadFile } from '@/utils';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
  Legend,
} from 'recharts';
import { format } from 'date-fns';
import { ArrowDownTrayIcon, ArrowPathIcon } from '@heroicons/react/24/outline';

// Mock data generators
const generateTrafficData = (hours: number) => {
  const data = [];
  const now = new Date();
  for (let i = hours; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 3600000);
    data.push({
      time: format(time, 'HH:mm'),
      in_bps: Math.random() * 8000000000 + 2000000000,
      out_bps: Math.random() * 4000000000 + 1000000000,
    });
  }
  return data;
};

const generateSystemData = (hours: number) => {
  const data = [];
  const now = new Date();
  for (let i = hours; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 3600000);
    data.push({
      time: format(time, 'HH:mm'),
      cpu: Math.random() * 40 + 30,
      memory: Math.random() * 20 + 60,
      temperature: Math.random() * 15 + 45,
    });
  }
  return data;
};

const timeRangeOptions = [
  { value: '1h', label: 'Last 1 hour' },
  { value: '6h', label: 'Last 6 hours' },
  { value: '24h', label: 'Last 24 hours' },
  { value: '7d', label: 'Last 7 days' },
  { value: '30d', label: 'Last 30 days' },
];

const routerOptions = [
  { value: '', label: 'All Routers' },
  { value: '1', label: 'Core-Router-1' },
  { value: '2', label: 'Edge-Router-1' },
  { value: '3', label: 'Distribution-Router-1' },
];

const interfaceOptions = [
  { value: '', label: 'All Interfaces' },
  { value: 'eth0', label: 'GigabitEthernet0/0' },
  { value: 'eth1', label: 'GigabitEthernet0/1' },
  { value: 'eth2', label: 'TenGigabitEthernet1/0' },
];

export function MetricsPage() {
  const [timeRange, setTimeRange] = useState<TimeRange>('24h');
  const [selectedRouter, setSelectedRouter] = useState('');
  const [selectedInterface, setSelectedInterface] = useState('');

  const getHoursFromRange = (range: TimeRange): number => {
    switch (range) {
      case '1h':
        return 1;
      case '6h':
        return 6;
      case '24h':
        return 24;
      case '7d':
        return 168;
      case '30d':
        return 720;
      default:
        return 24;
    }
  };

  const hours = getHoursFromRange(timeRange);
  const trafficData = generateTrafficData(Math.min(hours, 48));
  const systemData = generateSystemData(Math.min(hours, 48));

  const handleExport = async () => {
    // Create CSV content
    const headers = ['Time', 'Inbound (bps)', 'Outbound (bps)'];
    const rows = trafficData.map((d) => [d.time, d.in_bps.toFixed(0), d.out_bps.toFixed(0)]);
    const csv = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    downloadFile(blob, `metrics-${timeRange}.csv`);
  };

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Metrics</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Network performance and system metrics
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={handleExport}>
            <ArrowDownTrayIcon className="h-4 w-4 mr-1" />
            Export
          </Button>
          <Button variant="secondary" size="sm">
            <ArrowPathIcon className="h-4 w-4 mr-1" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card className="p-4">
        <div className="flex flex-wrap items-center gap-4">
          <div className="w-full sm:w-48">
            <Select
              label="Time Range"
              value={timeRange}
              onChange={(v) => setTimeRange(v as TimeRange)}
              options={timeRangeOptions}
            />
          </div>
          <div className="w-full sm:w-48">
            <Select
              label="Router"
              value={selectedRouter}
              onChange={setSelectedRouter}
              options={routerOptions}
            />
          </div>
          <div className="w-full sm:w-48">
            <Select
              label="Interface"
              value={selectedInterface}
              onChange={setSelectedInterface}
              options={interfaceOptions}
            />
          </div>
        </div>
      </Card>

      {/* Traffic chart */}
      <Card title="Network Traffic">
        <div className="h-80">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={trafficData}>
              <defs>
                <linearGradient id="colorIn" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#10B981" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#10B981" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="colorOut" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#3B82F6" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
              <XAxis dataKey="time" tick={{ fill: '#6B7280', fontSize: 12 }} />
              <YAxis tickFormatter={(value) => formatBps(value)} tick={{ fill: '#6B7280', fontSize: 12 }} />
              <Tooltip formatter={(value) => typeof value === 'number' ? formatBps(value) : value} />
              <Legend />
              <Area
                type="monotone"
                dataKey="in_bps"
                stroke="#10B981"
                fill="url(#colorIn)"
                name="Inbound"
              />
              <Area
                type="monotone"
                dataKey="out_bps"
                stroke="#3B82F6"
                fill="url(#colorOut)"
                name="Outbound"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* System metrics */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* CPU & Memory */}
        <Card title="CPU & Memory Usage">
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={systemData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
                <XAxis dataKey="time" tick={{ fill: '#6B7280', fontSize: 12 }} />
                <YAxis
                  domain={[0, 100]}
                  tickFormatter={(v) => `${v}%`}
                  tick={{ fill: '#6B7280', fontSize: 12 }}
                />
                <Tooltip formatter={(value) => typeof value === 'number' ? formatPercentage(value) : value} />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="cpu"
                  stroke="#F59E0B"
                  strokeWidth={2}
                  dot={false}
                  name="CPU"
                />
                <Line
                  type="monotone"
                  dataKey="memory"
                  stroke="#8B5CF6"
                  strokeWidth={2}
                  dot={false}
                  name="Memory"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>

        {/* Temperature */}
        <Card title="Temperature">
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={systemData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200 dark:stroke-gray-700" />
                <XAxis dataKey="time" tick={{ fill: '#6B7280', fontSize: 12 }} />
                <YAxis
                  domain={[30, 80]}
                  tickFormatter={(v) => `${v}°C`}
                  tick={{ fill: '#6B7280', fontSize: 12 }}
                />
                <Tooltip formatter={(value) => typeof value === 'number' ? `${value.toFixed(1)}°C` : value} />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="temperature"
                  stroke="#EF4444"
                  strokeWidth={2}
                  dot={false}
                  name="Temperature"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </Card>
      </div>

      {/* Interface stats */}
      <Card title="Interface Statistics">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead>
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                  Interface
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                  In Traffic
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                  Out Traffic
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                  In Errors
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                  Out Errors
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 dark:text-gray-400">
                  Utilization
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
              {[
                {
                  name: 'GigabitEthernet0/0',
                  in_bps: 4500000000,
                  out_bps: 2300000000,
                  in_errors: 12,
                  out_errors: 0,
                  utilization: 45,
                },
                {
                  name: 'GigabitEthernet0/1',
                  in_bps: 2100000000,
                  out_bps: 1800000000,
                  in_errors: 0,
                  out_errors: 0,
                  utilization: 21,
                },
                {
                  name: 'TenGigabitEthernet1/0',
                  in_bps: 8500000000,
                  out_bps: 6200000000,
                  in_errors: 3,
                  out_errors: 1,
                  utilization: 85,
                },
              ].map((iface) => (
                <tr key={iface.name}>
                  <td className="px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">
                    {iface.name}
                  </td>
                  <td className="px-4 py-3 text-sm text-green-600">{formatBps(iface.in_bps)}</td>
                  <td className="px-4 py-3 text-sm text-blue-600">{formatBps(iface.out_bps)}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{iface.in_errors}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">{iface.out_errors}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-20 rounded-full bg-gray-200 dark:bg-gray-700">
                        <div
                          className={`h-2 rounded-full ${
                            iface.utilization > 80
                              ? 'bg-red-500'
                              : iface.utilization > 60
                              ? 'bg-yellow-500'
                              : 'bg-green-500'
                          }`}
                          style={{ width: `${iface.utilization}%` }}
                        />
                      </div>
                      <span className="text-sm text-gray-500">{iface.utilization}%</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}

export default MetricsPage;

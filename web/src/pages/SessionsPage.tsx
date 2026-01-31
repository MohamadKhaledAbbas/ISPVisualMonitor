import { useState } from 'react';
import {
  MagnifyingGlassIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';
import {
  Card,
  Button,
  Table,
  type TableColumn,
  Badge,
} from '@/components/common';
import type { PPPoESession, DHCPLease, NATConnection } from '@/types';
import { formatBytes, formatDuration } from '@/utils';
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/react';
import { cn } from '@/utils';

// Mock data
const mockPPPoESessions: PPPoESession[] = [
  {
    id: '1',
    router_id: 'r1',
    username: 'user001',
    ip_address: '100.64.1.1',
    mac_address: 'AA:BB:CC:DD:EE:01',
    interface_name: 'pppoe-out1',
    uptime: 86400,
    rx_bytes: 1500000000,
    tx_bytes: 500000000,
    session_id: 'ppp1234',
    created_at: new Date(Date.now() - 86400000).toISOString(),
  },
  {
    id: '2',
    router_id: 'r1',
    username: 'user002',
    ip_address: '100.64.1.2',
    mac_address: 'AA:BB:CC:DD:EE:02',
    interface_name: 'pppoe-out2',
    uptime: 3600,
    rx_bytes: 250000000,
    tx_bytes: 75000000,
    session_id: 'ppp1235',
    created_at: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: '3',
    router_id: 'r2',
    username: 'enterprise01',
    ip_address: '100.64.2.1',
    mac_address: 'AA:BB:CC:DD:EE:03',
    interface_name: 'pppoe-out3',
    uptime: 604800,
    rx_bytes: 50000000000,
    tx_bytes: 10000000000,
    session_id: 'ppp1236',
    created_at: new Date(Date.now() - 604800000).toISOString(),
  },
];

const mockDHCPLeases: DHCPLease[] = [
  {
    id: '1',
    router_id: 'r1',
    ip_address: '192.168.1.100',
    mac_address: '11:22:33:44:55:66',
    hostname: 'workstation-01',
    lease_time: 86400,
    expires_at: new Date(Date.now() + 43200000).toISOString(),
    created_at: new Date(Date.now() - 43200000).toISOString(),
  },
  {
    id: '2',
    router_id: 'r1',
    ip_address: '192.168.1.101',
    mac_address: '11:22:33:44:55:67',
    hostname: 'printer-01',
    lease_time: 86400,
    expires_at: new Date(Date.now() + 82800000).toISOString(),
    created_at: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: '3',
    router_id: 'r2',
    ip_address: '192.168.2.50',
    mac_address: '11:22:33:44:55:68',
    lease_time: 3600,
    expires_at: new Date(Date.now() + 1800000).toISOString(),
    created_at: new Date(Date.now() - 1800000).toISOString(),
  },
];

const mockNATConnections: NATConnection[] = [
  {
    id: '1',
    router_id: 'r1',
    protocol: 'tcp',
    src_address: '192.168.1.100',
    dst_address: '8.8.8.8',
    src_port: 45678,
    dst_port: 443,
    reply_src_address: '8.8.8.8',
    reply_dst_address: '100.64.1.1',
    reply_src_port: 443,
    reply_dst_port: 45678,
    state: 'ESTABLISHED',
    timeout: 432000,
    created_at: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: '2',
    router_id: 'r1',
    protocol: 'udp',
    src_address: '192.168.1.101',
    dst_address: '1.1.1.1',
    src_port: 12345,
    dst_port: 53,
    state: 'UNREPLIED',
    timeout: 30,
    created_at: new Date(Date.now() - 10000).toISOString(),
  },
];

export function SessionsPage() {
  const [search, setSearch] = useState('');

  // Filter mock data based on search
  const filteredPPPoE = mockPPPoESessions.filter(
    (s) =>
      !search ||
      s.username.toLowerCase().includes(search.toLowerCase()) ||
      s.ip_address.includes(search) ||
      s.mac_address.toLowerCase().includes(search.toLowerCase())
  );

  const filteredDHCP = mockDHCPLeases.filter(
    (s) =>
      !search ||
      s.ip_address.includes(search) ||
      s.mac_address.toLowerCase().includes(search.toLowerCase()) ||
      s.hostname?.toLowerCase().includes(search.toLowerCase())
  );

  const filteredNAT = mockNATConnections.filter(
    (s) =>
      !search ||
      s.src_address.includes(search) ||
      s.dst_address.includes(search)
  );

  const pppoeColumns: TableColumn<PPPoESession>[] = [
    {
      key: 'username',
      header: 'Username',
      cell: (session) => (
        <span className="font-medium text-gray-900 dark:text-white">{session.username}</span>
      ),
    },
    {
      key: 'ip',
      header: 'IP Address',
      cell: (session) => <span className="font-mono text-sm">{session.ip_address}</span>,
    },
    {
      key: 'mac',
      header: 'MAC Address',
      cell: (session) => (
        <span className="font-mono text-sm text-gray-500">{session.mac_address}</span>
      ),
    },
    {
      key: 'uptime',
      header: 'Uptime',
      cell: (session) => formatDuration(session.uptime),
    },
    {
      key: 'traffic',
      header: 'Traffic (RX/TX)',
      cell: (session) => (
        <div className="text-sm">
          <span className="text-green-600">{formatBytes(session.rx_bytes)}</span>
          {' / '}
          <span className="text-blue-600">{formatBytes(session.tx_bytes)}</span>
        </div>
      ),
    },
    {
      key: 'interface',
      header: 'Interface',
      cell: (session) => session.interface_name,
    },
  ];

  const dhcpColumns: TableColumn<DHCPLease>[] = [
    {
      key: 'ip',
      header: 'IP Address',
      cell: (lease) => <span className="font-mono text-sm">{lease.ip_address}</span>,
    },
    {
      key: 'mac',
      header: 'MAC Address',
      cell: (lease) => <span className="font-mono text-sm text-gray-500">{lease.mac_address}</span>,
    },
    {
      key: 'hostname',
      header: 'Hostname',
      cell: (lease) => lease.hostname || '-',
    },
    {
      key: 'expires',
      header: 'Expires In',
      cell: (lease) => {
        const expiresIn = Math.max(0, (new Date(lease.expires_at).getTime() - Date.now()) / 1000);
        return formatDuration(expiresIn);
      },
    },
    {
      key: 'status',
      header: 'Status',
      cell: (lease) => {
        const expiresIn = new Date(lease.expires_at).getTime() - Date.now();
        return (
          <Badge variant={expiresIn > 3600000 ? 'success' : expiresIn > 0 ? 'warning' : 'danger'}>
            {expiresIn > 0 ? 'Active' : 'Expired'}
          </Badge>
        );
      },
    },
  ];

  const natColumns: TableColumn<NATConnection>[] = [
    {
      key: 'protocol',
      header: 'Protocol',
      cell: (conn) => (
        <Badge variant="default" size="sm">
          {conn.protocol.toUpperCase()}
        </Badge>
      ),
    },
    {
      key: 'source',
      header: 'Source',
      cell: (conn) => (
        <span className="font-mono text-sm">
          {conn.src_address}
          {conn.src_port && `:${conn.src_port}`}
        </span>
      ),
    },
    {
      key: 'destination',
      header: 'Destination',
      cell: (conn) => (
        <span className="font-mono text-sm">
          {conn.dst_address}
          {conn.dst_port && `:${conn.dst_port}`}
        </span>
      ),
    },
    {
      key: 'state',
      header: 'State',
      cell: (conn) => (
        <Badge
          variant={conn.state === 'ESTABLISHED' ? 'success' : conn.state === 'UNREPLIED' ? 'warning' : 'default'}
          size="sm"
        >
          {conn.state}
        </Badge>
      ),
    },
    {
      key: 'timeout',
      header: 'Timeout',
      cell: (conn) => formatDuration(conn.timeout),
    },
  ];

  const tabs = ['PPPoE Sessions', 'DHCP Leases', 'NAT Connections'];

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Sessions</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Monitor active network sessions
          </p>
        </div>
        <Button variant="secondary">
          <ArrowPathIcon className="h-4 w-4 mr-1" />
          Refresh
        </Button>
      </div>

      {/* Search */}
      <Card className="p-4">
        <div className="relative max-w-md">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search by username, IP, MAC, or hostname..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-md border border-gray-300 bg-white py-2 pl-10 pr-4 text-sm placeholder:text-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
          />
        </div>
      </Card>

      {/* Tabs */}
      <TabGroup>
        <TabList className="flex space-x-1 rounded-lg bg-gray-100 p-1 dark:bg-gray-800">
          {tabs.map((tab) => (
            <Tab
              key={tab}
              className={({ selected }) =>
                cn(
                  'w-full rounded-md py-2.5 text-sm font-medium leading-5 transition-colors',
                  'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
                  selected
                    ? 'bg-white text-blue-700 shadow dark:bg-gray-700 dark:text-blue-400'
                    : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-white'
                )
              }
            >
              {tab}
            </Tab>
          ))}
        </TabList>
        <TabPanels className="mt-4">
          <TabPanel>
            <Card padding="none">
              <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {filteredPPPoE.length} active PPPoE sessions
                </p>
              </div>
              <Table
                columns={pppoeColumns}
                data={filteredPPPoE}
                keyExtractor={(s) => s.id}
                emptyMessage="No PPPoE sessions found"
              />
            </Card>
          </TabPanel>
          <TabPanel>
            <Card padding="none">
              <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {filteredDHCP.length} active DHCP leases
                </p>
              </div>
              <Table
                columns={dhcpColumns}
                data={filteredDHCP}
                keyExtractor={(s) => s.id}
                emptyMessage="No DHCP leases found"
              />
            </Card>
          </TabPanel>
          <TabPanel>
            <Card padding="none">
              <div className="p-4 border-b border-gray-200 dark:border-gray-700">
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {filteredNAT.length} NAT connections
                </p>
              </div>
              <Table
                columns={natColumns}
                data={filteredNAT}
                keyExtractor={(s) => s.id}
                emptyMessage="No NAT connections found"
              />
            </Card>
          </TabPanel>
        </TabPanels>
      </TabGroup>
    </div>
  );
}

export default SessionsPage;

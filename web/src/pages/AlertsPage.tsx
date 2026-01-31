import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  FunnelIcon,
  CheckIcon,
  BellIcon,
} from '@heroicons/react/24/outline';
import {
  Card,
  Button,
  Table,
  type TableColumn,
  Pagination,
  Badge,
  Select,
  Modal,
  Input,
} from '@/components/common';
import { alertsApi } from '@/api';
import type { Alert, AlertSeverity, AlertStatus } from '@/types';
import { formatDuration } from '@/utils';

// Mock data
const mockAlerts: Alert[] = [
  {
    id: '1',
    tenant_id: '1',
    router_id: 'r1',
    severity: 'critical',
    status: 'open',
    title: 'Router Core-Router-1 offline',
    description: 'Router is not responding to SNMP polls for more than 5 minutes',
    created_at: new Date(Date.now() - 300000).toISOString(),
    updated_at: new Date(Date.now() - 300000).toISOString(),
  },
  {
    id: '2',
    tenant_id: '1',
    router_id: 'r2',
    severity: 'warning',
    status: 'open',
    title: 'High CPU usage on Edge-Router-1',
    description: 'CPU utilization has been above 90% for the last 15 minutes',
    created_at: new Date(Date.now() - 900000).toISOString(),
    updated_at: new Date(Date.now() - 900000).toISOString(),
  },
  {
    id: '3',
    tenant_id: '1',
    interface_id: 'if1',
    severity: 'warning',
    status: 'acknowledged',
    title: 'Interface flapping on Distribution-Router-1',
    description: 'GigabitEthernet0/1 status changed 5 times in the last hour',
    acknowledged_by: 'admin@example.com',
    acknowledged_at: new Date(Date.now() - 600000).toISOString(),
    acknowledged_note: 'Investigating with NOC team',
    created_at: new Date(Date.now() - 3600000).toISOString(),
    updated_at: new Date(Date.now() - 600000).toISOString(),
  },
  {
    id: '4',
    tenant_id: '1',
    router_id: 'r3',
    severity: 'info',
    status: 'resolved',
    title: 'Configuration change detected on Access-Router-1',
    description: 'Router configuration was modified',
    resolved_at: new Date(Date.now() - 1800000).toISOString(),
    created_at: new Date(Date.now() - 7200000).toISOString(),
    updated_at: new Date(Date.now() - 1800000).toISOString(),
  },
  {
    id: '5',
    tenant_id: '1',
    router_id: 'r4',
    severity: 'critical',
    status: 'open',
    title: 'Memory usage critical on Core-Router-2',
    description: 'Memory usage exceeded 95% threshold',
    created_at: new Date(Date.now() - 120000).toISOString(),
    updated_at: new Date(Date.now() - 120000).toISOString(),
  },
];

const severityOptions = [
  { value: '', label: 'All Severities' },
  { value: 'critical', label: 'Critical' },
  { value: 'warning', label: 'Warning' },
  { value: 'info', label: 'Info' },
];

const statusOptions = [
  { value: '', label: 'All Statuses' },
  { value: 'open', label: 'Open' },
  { value: 'acknowledged', label: 'Acknowledged' },
  { value: 'resolved', label: 'Resolved' },
];

export function AlertsPage() {
  const queryClient = useQueryClient();
  const [severityFilter, setSeverityFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);
  const [showAckModal, setShowAckModal] = useState(false);
  const [alertToAck, setAlertToAck] = useState<Alert | null>(null);
  const [ackNote, setAckNote] = useState('');

  // In production, fetch from API
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['alerts', { page, severity: severityFilter, status: statusFilter }],
    queryFn: () =>
      alertsApi.list({
        page,
        page_size: 20,
        severity: severityFilter as AlertSeverity || undefined,
        status: statusFilter as AlertStatus || undefined,
      }),
    enabled: false, // Using mock data
  });

  // Filter mock data
  const filteredAlerts = mockAlerts.filter((alert) => {
    const matchesSeverity = !severityFilter || alert.severity === severityFilter;
    const matchesStatus = !statusFilter || alert.status === statusFilter;
    return matchesSeverity && matchesStatus;
  });

  const alerts = data?.data || filteredAlerts;
  const pagination = data?.pagination || {
    page: 1,
    page_size: 20,
    total_items: filteredAlerts.length,
    total_pages: 1,
  };

  useMutation({
    mutationFn: ({ id, note }: { id: string; note?: string }) =>
      alertsApi.acknowledge(id, { note }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      setShowAckModal(false);
      setAlertToAck(null);
      setAckNote('');
    },
  });

  const handleAcknowledge = (alert: Alert) => {
    setAlertToAck(alert);
    setShowAckModal(true);
  };

  const confirmAcknowledge = () => {
    if (alertToAck) {
      // For now just close the modal
      setShowAckModal(false);
      setAlertToAck(null);
      setAckNote('');
    }
  };

  const getSeverityVariant = (severity: AlertSeverity): 'danger' | 'warning' | 'info' => {
    switch (severity) {
      case 'critical':
        return 'danger';
      case 'warning':
        return 'warning';
      default:
        return 'info';
    }
  };

  const getStatusVariant = (status: AlertStatus): 'danger' | 'warning' | 'success' | 'default' => {
    switch (status) {
      case 'open':
        return 'danger';
      case 'acknowledged':
        return 'warning';
      case 'resolved':
        return 'success';
      default:
        return 'default';
    }
  };

  const columns: TableColumn<Alert>[] = [
    {
      key: 'severity',
      header: 'Severity',
      cell: (alert) => (
        <Badge variant={getSeverityVariant(alert.severity)} size="sm">
          {alert.severity}
        </Badge>
      ),
      className: 'w-24',
    },
    {
      key: 'title',
      header: 'Alert',
      cell: (alert) => (
        <div>
          <p className="font-medium text-gray-900 dark:text-white">{alert.title}</p>
          <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-1">
            {alert.description}
          </p>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      cell: (alert) => (
        <Badge variant={getStatusVariant(alert.status)} size="sm">
          {alert.status}
        </Badge>
      ),
      className: 'w-32',
    },
    {
      key: 'created',
      header: 'Created',
      cell: (alert) => (
        <span className="text-sm text-gray-500 dark:text-gray-400">
          {formatDuration((Date.now() - new Date(alert.created_at).getTime()) / 1000)} ago
        </span>
      ),
      className: 'w-32',
    },
    {
      key: 'actions',
      header: '',
      cell: (alert) =>
        alert.status === 'open' ? (
          <Button
            variant="secondary"
            size="sm"
            onClick={(e) => {
              e.stopPropagation();
              handleAcknowledge(alert);
            }}
          >
            <CheckIcon className="h-4 w-4 mr-1" />
            Ack
          </Button>
        ) : null,
      className: 'w-24',
    },
  ];

  // Stats
  const openCount = alerts.filter((a) => a.status === 'open').length;
  const criticalCount = alerts.filter((a) => a.severity === 'critical' && a.status === 'open').length;
  const warningCount = alerts.filter((a) => a.severity === 'warning' && a.status === 'open').length;

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Alerts</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Monitor and manage system alerts
          </p>
        </div>
        <Button variant="secondary" onClick={() => refetch()}>
          Refresh
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-3">
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Open Alerts</p>
              <p className="text-2xl font-semibold text-gray-900 dark:text-white">{openCount}</p>
            </div>
            <div className="rounded-full bg-blue-100 p-2 dark:bg-blue-900/20">
              <BellIcon className="h-6 w-6 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </Card>
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Critical</p>
              <p className="text-2xl font-semibold text-red-600">{criticalCount}</p>
            </div>
            <div className="rounded-full bg-red-100 p-2 dark:bg-red-900/20">
              <BellIcon className="h-6 w-6 text-red-600" />
            </div>
          </div>
        </Card>
        <Card className="p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Warning</p>
              <p className="text-2xl font-semibold text-yellow-600">{warningCount}</p>
            </div>
            <div className="rounded-full bg-yellow-100 p-2 dark:bg-yellow-900/20">
              <BellIcon className="h-6 w-6 text-yellow-600" />
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card className="p-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <FunnelIcon className="hidden h-5 w-5 text-gray-400 sm:block" />
          <div className="w-full sm:w-48">
            <Select
              value={severityFilter}
              onChange={setSeverityFilter}
              options={severityOptions}
              placeholder="Severity"
            />
          </div>
          <div className="w-full sm:w-48">
            <Select
              value={statusFilter}
              onChange={setStatusFilter}
              options={statusOptions}
              placeholder="Status"
            />
          </div>
        </div>
      </Card>

      {/* Alerts table */}
      <Card padding="none">
        <Table
          columns={columns}
          data={alerts}
          keyExtractor={(alert) => alert.id}
          isLoading={isLoading}
          emptyMessage="No alerts found"
        />
        {pagination.total_pages > 1 && (
          <Pagination
            currentPage={pagination.page}
            totalPages={pagination.total_pages}
            totalItems={pagination.total_items}
            pageSize={pagination.page_size}
            onPageChange={setPage}
          />
        )}
      </Card>

      {/* Acknowledge modal */}
      <Modal
        isOpen={showAckModal}
        onClose={() => setShowAckModal(false)}
        title="Acknowledge Alert"
        size="md"
      >
        {alertToAck && (
          <div className="space-y-4">
            <div>
              <h4 className="font-medium text-gray-900 dark:text-white">{alertToAck.title}</h4>
              <p className="text-sm text-gray-500 dark:text-gray-400">{alertToAck.description}</p>
            </div>
            <Input
              label="Note (optional)"
              value={ackNote}
              onChange={(e) => setAckNote(e.target.value)}
              placeholder="Add a note about this acknowledgement..."
            />
            <div className="flex justify-end gap-2">
              <Button variant="secondary" onClick={() => setShowAckModal(false)}>
                Cancel
              </Button>
              <Button onClick={confirmAcknowledge}>
                <CheckIcon className="h-4 w-4 mr-1" />
                Acknowledge
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}

export default AlertsPage;

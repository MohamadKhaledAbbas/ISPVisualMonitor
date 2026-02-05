import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import {
  PlusIcon,
  MagnifyingGlassIcon,
  TrashIcon,
  ArrowPathIcon,
  EllipsisVerticalIcon,
} from '@heroicons/react/24/outline';
import {
  Card,
  Button,
  Table,
  type TableColumn,
  Pagination,
  Badge,
  StatusBadge,
  Select,
  Modal,
} from '@/components/common';
import { routersApi } from '@/api';
import type { Router } from '@/types';
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/react';

// Mock data for development
const mockRouters: Router[] = [
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
    location: { latitude: 40.7128, longitude: -74.006 },
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
    location: { latitude: 34.0522, longitude: -118.2437 },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '3',
    tenant_id: '1',
    name: 'Distribution-Router-1',
    hostname: 'dist1.example.com',
    management_ip: '10.0.0.3',
    vendor: 'Juniper',
    model: 'MX480',
    status: 'offline',
    polling_enabled: true,
    location: { latitude: 41.8781, longitude: -87.6298 },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '4',
    tenant_id: '1',
    name: 'Access-Router-1',
    hostname: 'access1.example.com',
    management_ip: '10.0.0.4',
    vendor: 'MikroTik',
    model: 'CCR2004',
    status: 'maintenance',
    polling_enabled: false,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '5',
    tenant_id: '1',
    name: 'Edge-Router-2',
    hostname: 'edge2.example.com',
    management_ip: '10.0.0.5',
    vendor: 'Cisco',
    model: 'ISR4321',
    status: 'active',
    polling_enabled: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

const statusOptions = [
  { value: '', label: 'All Statuses' },
  { value: 'active', label: 'Active' },
  { value: 'offline', label: 'Offline' },
  { value: 'maintenance', label: 'Maintenance' },
];

export function RoutersPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [page, setPage] = useState(1);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [routerToDelete, setRouterToDelete] = useState<string | null>(null);

  // In production, fetch from API
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['routers', { page, search, status: statusFilter }],
    queryFn: () => routersApi.list({ page, page_size: 20, search, status: statusFilter }),
    enabled: false, // Using mock data
  });

  // Filter mock data
  const filteredRouters = mockRouters.filter((router) => {
    const matchesSearch =
      !search ||
      router.name.toLowerCase().includes(search.toLowerCase()) ||
      router.hostname.toLowerCase().includes(search.toLowerCase()) ||
      router.management_ip.includes(search);
    const matchesStatus = !statusFilter || router.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const routers = data?.data || filteredRouters;
  const pagination = data?.pagination || {
    page: 1,
    page_size: 20,
    total_items: filteredRouters.length,
    total_pages: 1,
  };

  useMutation({
    mutationFn: (id: string) => routersApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['routers'] });
      setShowDeleteModal(false);
      setRouterToDelete(null);
    },
  });

  const handleDelete = (id: string) => {
    setRouterToDelete(id);
    setShowDeleteModal(true);
  };

  const confirmDelete = () => {
    if (routerToDelete) {
      // For now just close the modal
      setShowDeleteModal(false);
      setRouterToDelete(null);
    }
  };

  const columns: TableColumn<Router>[] = [
    {
      key: 'name',
      header: 'Name',
      cell: (router) => (
        <div>
          <Link
            to={`/routers/${router.id}`}
            className="font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400"
          >
            {router.name}
          </Link>
          <p className="text-xs text-gray-500 dark:text-gray-400">{router.hostname}</p>
        </div>
      ),
    },
    {
      key: 'ip',
      header: 'Management IP',
      cell: (router) => (
        <span className="font-mono text-sm">{router.management_ip}</span>
      ),
    },
    {
      key: 'vendor',
      header: 'Vendor / Model',
      cell: (router) => (
        <div>
          <p className="text-sm">{router.vendor || '-'}</p>
          <p className="text-xs text-gray-500 dark:text-gray-400">{router.model || '-'}</p>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      cell: (router) => <StatusBadge status={router.status} />,
    },
    {
      key: 'polling',
      header: 'Polling',
      cell: (router) => (
        <Badge variant={router.polling_enabled ? 'success' : 'default'} size="sm">
          {router.polling_enabled ? 'Enabled' : 'Disabled'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      header: '',
      cell: (router) => (
        <Menu as="div" className="relative">
          <MenuButton className="p-1 text-gray-400 hover:text-gray-500">
            <EllipsisVerticalIcon className="h-5 w-5" />
          </MenuButton>
          <MenuItems className="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none dark:bg-gray-800">
            <MenuItem>
              {({ active }) => (
                <Link
                  to={`/routers/${router.id}`}
                  className={`block px-4 py-2 text-sm ${
                    active ? 'bg-gray-100 dark:bg-gray-700' : ''
                  } text-gray-700 dark:text-gray-200`}
                >
                  View Details
                </Link>
              )}
            </MenuItem>
            <MenuItem>
              {({ active }) => (
                <Link
                  to={`/routers/${router.id}/edit`}
                  className={`block px-4 py-2 text-sm ${
                    active ? 'bg-gray-100 dark:bg-gray-700' : ''
                  } text-gray-700 dark:text-gray-200`}
                >
                  Edit
                </Link>
              )}
            </MenuItem>
            <MenuItem>
              {({ active }) => (
                <button
                  onClick={() => handleDelete(router.id)}
                  className={`block w-full px-4 py-2 text-left text-sm ${
                    active ? 'bg-gray-100 dark:bg-gray-700' : ''
                  } text-red-600 dark:text-red-400`}
                >
                  Delete
                </button>
              )}
            </MenuItem>
          </MenuItems>
        </Menu>
      ),
      className: 'w-10',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Routers</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Manage your network infrastructure
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={() => refetch()}>
            <ArrowPathIcon className="h-4 w-4 mr-1" />
            Refresh
          </Button>
          <Button onClick={() => navigate('/routers/new')}>
            <PlusIcon className="h-4 w-4 mr-1" />
            Add Router
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card className="p-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div className="relative flex-1">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Search by name, hostname, or IP..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full rounded-md border border-gray-300 bg-white py-2 pl-10 pr-4 text-sm placeholder:text-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
            />
          </div>
          <div className="w-full sm:w-48">
            <Select
              value={statusFilter}
              onChange={setStatusFilter}
              options={statusOptions}
              placeholder="Filter by status"
            />
          </div>
          {selectedIds.size > 0 && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">
                {selectedIds.size} selected
              </span>
              <Button
                variant="danger"
                size="sm"
                onClick={() => setShowDeleteModal(true)}
              >
                <TrashIcon className="h-4 w-4" />
              </Button>
            </div>
          )}
        </div>
      </Card>

      {/* Router table */}
      <Card padding="none">
        <Table
          columns={columns}
          data={routers}
          keyExtractor={(router) => router.id}
          onRowClick={(router) => navigate(`/routers/${router.id}`)}
          selectedKeys={selectedIds}
          onSelectionChange={setSelectedIds}
          isLoading={isLoading}
          emptyMessage="No routers found"
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

      {/* Delete confirmation modal */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Delete Router"
        size="sm"
      >
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Are you sure you want to delete this router? This action cannot be undone.
        </p>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="secondary" onClick={() => setShowDeleteModal(false)}>
            Cancel
          </Button>
          <Button variant="danger" onClick={confirmDelete}>
            Delete
          </Button>
        </div>
      </Modal>
    </div>
  );
}

export default RoutersPage;

import { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import maplibregl, { Map, Marker } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { Card, Badge, Button } from '@/components/common';
import { cn, formatBps } from '@/utils';
import {
  EyeIcon,
  EyeSlashIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  SignalIcon,
  UserGroupIcon,
  CpuChipIcon,
  ServerStackIcon,
  WifiIcon,
  BellAlertIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  XMarkIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';

// ─── Types ──────────────────────────────────────────────────────────────────

interface RouterNode {
  id: string;
  name: string;
  role: 'core' | 'distribution' | 'access' | 'pppoe';
  coordinates: [number, number];
  status: 'active' | 'degraded' | 'offline' | 'maintenance';
  vendor: string;
  model: string;
  management_ip: string;
  cpu_usage: number;
  memory_usage: number;
  temperature: number;
  uptime_seconds: number;
  pppoe_sessions: number;
  max_sessions: number;
  bandwidth_in: number;
  bandwidth_out: number;
  interface_count: number;
  active_alerts: ProactiveAlert[];
}

interface NetworkLink {
  id: string;
  source_id: string;
  target_id: string;
  status: 'active' | 'degraded' | 'down';
  utilization: number;
  bandwidth_capacity: number;
  bandwidth_used: number;
  latency_ms: number;
  packet_loss: number;
}

interface PPPoEUser {
  id: string;
  username: string;
  ip_address: string;
  router_id: string;
  rx_rate: number;
  tx_rate: number;
  plan: string;
  status: 'connected' | 'authenticating';
}

interface ProactiveAlert {
  id: string;
  router_id: string;
  severity: 'critical' | 'warning' | 'info';
  title: string;
  description: string;
  impact: string;
  affected_users: number;
  detected_at: number;
  auto_detected: boolean;
}

// ─── Simulation Data ────────────────────────────────────────────────────────

const ISP_PLANS = ['Basic-10M', 'Standard-25M', 'Premium-50M', 'Business-100M', 'Enterprise-200M'];

function generateSimulationData() {
  const routers: RouterNode[] = [
    {
      id: 'core-01', name: 'CHR-Core-01', role: 'core',
      coordinates: [35.5018, 33.8938], status: 'active',
      vendor: 'MikroTik', model: 'CHR', management_ip: '10.0.0.1',
      cpu_usage: 35 + Math.random() * 15, memory_usage: 58 + Math.random() * 10,
      temperature: 42 + Math.random() * 8, uptime_seconds: 2592000 + Math.floor(Math.random() * 604800),
      pppoe_sessions: 0, max_sessions: 0,
      bandwidth_in: 4.5e9 + Math.random() * 2e9, bandwidth_out: 2.3e9 + Math.random() * 1e9,
      interface_count: 12, active_alerts: [],
    },
    {
      id: 'edge-01', name: 'CHR-Edge-01', role: 'distribution',
      coordinates: [35.8308, 33.8547], status: 'active',
      vendor: 'MikroTik', model: 'CHR', management_ip: '10.0.1.1',
      cpu_usage: 42 + Math.random() * 20, memory_usage: 65 + Math.random() * 10,
      temperature: 45 + Math.random() * 10, uptime_seconds: 1296000 + Math.floor(Math.random() * 604800),
      pppoe_sessions: 0, max_sessions: 0,
      bandwidth_in: 2.1e9 + Math.random() * 5e8, bandwidth_out: 1.1e9 + Math.random() * 3e8,
      interface_count: 8, active_alerts: [],
    },
    {
      id: 'access-01', name: 'CHR-Access-01', role: 'access',
      coordinates: [35.5437, 33.8886], status: 'active',
      vendor: 'MikroTik', model: 'CHR', management_ip: '10.0.2.1',
      cpu_usage: 55 + Math.random() * 20, memory_usage: 70 + Math.random() * 15,
      temperature: 48 + Math.random() * 12, uptime_seconds: 864000 + Math.floor(Math.random() * 604800),
      pppoe_sessions: 0, max_sessions: 500,
      bandwidth_in: 8e8 + Math.random() * 2e8, bandwidth_out: 4e8 + Math.random() * 1e8,
      interface_count: 6, active_alerts: [],
    },
    {
      id: 'pppoe-01', name: 'CHR-PPPoE-01', role: 'pppoe',
      coordinates: [35.4955, 33.8731], status: 'degraded',
      vendor: 'MikroTik', model: 'CHR', management_ip: '10.0.3.1',
      cpu_usage: 82 + Math.random() * 10, memory_usage: 88 + Math.random() * 8,
      temperature: 62 + Math.random() * 8, uptime_seconds: 432000,
      pppoe_sessions: 245, max_sessions: 300,
      bandwidth_in: 1.2e9 + Math.random() * 3e8, bandwidth_out: 6e8 + Math.random() * 1.5e8,
      interface_count: 4, active_alerts: [],
    },
    {
      id: 'pppoe-02', name: 'CHR-PPPoE-02', role: 'pppoe',
      coordinates: [35.5650, 33.8600], status: 'active',
      vendor: 'MikroTik', model: 'CHR', management_ip: '10.0.4.1',
      cpu_usage: 38 + Math.random() * 15, memory_usage: 52 + Math.random() * 10,
      temperature: 44 + Math.random() * 6, uptime_seconds: 1728000,
      pppoe_sessions: 178, max_sessions: 300,
      bandwidth_in: 9e8 + Math.random() * 2e8, bandwidth_out: 4.5e8 + Math.random() * 1e8,
      interface_count: 4, active_alerts: [],
    },
    {
      id: 'pppoe-03', name: 'CHR-PPPoE-03', role: 'pppoe',
      coordinates: [35.8600, 33.8800], status: 'offline',
      vendor: 'MikroTik', model: 'CHR', management_ip: '10.0.5.1',
      cpu_usage: 0, memory_usage: 0, temperature: 0, uptime_seconds: 0,
      pppoe_sessions: 0, max_sessions: 300,
      bandwidth_in: 0, bandwidth_out: 0,
      interface_count: 4, active_alerts: [],
    },
  ];

  const alerts: ProactiveAlert[] = [
    {
      id: 'a1', router_id: 'pppoe-01', severity: 'warning',
      title: 'High CPU Utilization',
      description: 'CPU usage above 80% for 15 min. PPPoE processing may be delayed.',
      impact: 'Potential session drops and slow authentication',
      affected_users: 245, detected_at: Date.now() - 900_000, auto_detected: true,
    },
    {
      id: 'a2', router_id: 'pppoe-01', severity: 'warning',
      title: 'Memory Pressure Detected',
      description: 'Memory at 88%. Approaching critical 90% threshold.',
      impact: 'Risk of OOM-related session terminations',
      affected_users: 245, detected_at: Date.now() - 600_000, auto_detected: true,
    },
    {
      id: 'a3', router_id: 'pppoe-03', severity: 'critical',
      title: 'Router Unreachable',
      description: 'CHR-PPPoE-03 not responding for 8 min. Last state: 156 sessions active.',
      impact: '156 PPPoE users disconnected. Kaslik area offline.',
      affected_users: 156, detected_at: Date.now() - 480_000, auto_detected: true,
    },
    {
      id: 'a4', router_id: 'edge-01', severity: 'info',
      title: 'Scheduled Maintenance',
      description: 'Firmware upgrade scheduled. ~5 min downtime expected.',
      impact: 'Brief interruption for users routed through Edge-01',
      affected_users: 423, detected_at: Date.now() - 3600_000, auto_detected: false,
    },
  ];

  alerts.forEach(a => {
    const r = routers.find(x => x.id === a.router_id);
    if (r) r.active_alerts.push(a);
  });

  const links: NetworkLink[] = [
    { id: 'l1', source_id: 'core-01', target_id: 'edge-01', status: 'active', utilization: 62, bandwidth_capacity: 1e10, bandwidth_used: 6.2e9, latency_ms: 1.2, packet_loss: 0 },
    { id: 'l2', source_id: 'core-01', target_id: 'access-01', status: 'active', utilization: 45, bandwidth_capacity: 1e10, bandwidth_used: 4.5e9, latency_ms: 0.8, packet_loss: 0 },
    { id: 'l3', source_id: 'access-01', target_id: 'pppoe-01', status: 'degraded', utilization: 85, bandwidth_capacity: 1e9, bandwidth_used: 8.5e8, latency_ms: 8.5, packet_loss: 0.3 },
    { id: 'l4', source_id: 'access-01', target_id: 'pppoe-02', status: 'active', utilization: 52, bandwidth_capacity: 1e9, bandwidth_used: 5.2e8, latency_ms: 2.1, packet_loss: 0 },
    { id: 'l5', source_id: 'edge-01', target_id: 'pppoe-03', status: 'down', utilization: 0, bandwidth_capacity: 1e9, bandwidth_used: 0, latency_ms: 999, packet_loss: 100 },
  ];

  const users: PPPoEUser[] = [];
  routers.filter(r => r.role === 'pppoe' && r.status !== 'offline').forEach(router => {
    for (let i = 0; i < router.pppoe_sessions; i++) {
      users.push({
        id: `${router.id}-u${i}`,
        username: `user${String(i + 1).padStart(4, '0')}@isp.local`,
        ip_address: `100.64.${router.id === 'pppoe-01' ? 1 : 2}.${(i % 254) + 1}`,
        router_id: router.id,
        rx_rate: Math.floor(Math.random() * 5e7),
        tx_rate: Math.floor(Math.random() * 2.5e7),
        plan: ISP_PLANS[Math.floor(Math.random() * ISP_PLANS.length)],
        status: Math.random() > 0.02 ? 'connected' : 'authenticating',
      });
    }
  });

  return { routers, links, users, alerts };
}

// ─── Helpers ────────────────────────────────────────────────────────────────

const statusColors: Record<string, string> = {
  active: '#10B981', online: '#10B981', connected: '#10B981',
  offline: '#EF4444', down: '#EF4444',
  maintenance: '#F59E0B', degraded: '#F59E0B',
};

function formatUptime(s: number): string {
  if (!s) return 'N/A';
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  return d > 0 ? `${d}d ${h}h` : `${h}h ${Math.floor((s % 3600) / 60)}m`;
}

function sevColor(s: string) {
  return s === 'critical' ? 'text-red-600 dark:text-red-400' : s === 'warning' ? 'text-amber-600 dark:text-amber-400' : 'text-blue-600 dark:text-blue-400';
}
function sevBg(s: string) {
  return s === 'critical' ? 'bg-red-50 border-red-200 dark:bg-red-950/30 dark:border-red-800' : s === 'warning' ? 'bg-amber-50 border-amber-200 dark:bg-amber-950/30 dark:border-amber-800' : 'bg-blue-50 border-blue-200 dark:bg-blue-950/30 dark:border-blue-800';
}
function nodeSize(role: string) { return role === 'core' ? 40 : role === 'distribution' ? 32 : role === 'access' ? 28 : 24; }
function linkColor(l: NetworkLink) { return l.status === 'down' ? '#EF4444' : l.utilization > 80 ? '#F59E0B' : l.utilization > 60 ? '#3B82F6' : '#10B981'; }
function linkWidth(u: number) { return u === 0 ? 2 : u < 50 ? 3 : u < 80 ? 4 : 5; }

// ─── Main Component ─────────────────────────────────────────────────────────

export function MapPage() {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<Map | null>(null);
  const markersRef = useRef<Marker[]>([]);

  const [selectedNode, setSelectedNode] = useState<RouterNode | null>(null);
  const [alertsOpen, setAlertsOpen] = useState(true);
  const [usersOpen, setUsersOpen] = useState(false);
  const [layers, setLayers] = useState({ routers: true, links: true, users: true });
  const [tick, setTick] = useState(Date.now());
  const [spinning, setSpinning] = useState(false);

  const sim = useMemo(() => generateSimulationData(), [tick]);
  const totalUsers = sim.users.length;
  const critAlerts = sim.alerts.filter(a => a.severity === 'critical').length;
  const bwIn = sim.routers.reduce((s, r) => s + r.bandwidth_in, 0);
  const bwOut = sim.routers.reduce((s, r) => s + r.bandwidth_out, 0);
  const downCount = sim.routers.filter(r => r.status === 'offline').length;

  const refresh = useCallback(() => {
    setSpinning(true);
    setTimeout(() => { setTick(Date.now()); setSpinning(false); }, 500);
  }, []);

  useEffect(() => {
    const iv = setInterval(() => setTick(Date.now()), 30_000);
    return () => clearInterval(iv);
  }, []);

  // ── Map init ──
  useEffect(() => {
    if (!mapContainer.current || mapRef.current) return;
    mapRef.current = new maplibregl.Map({
      container: mapContainer.current,
      style: {
        version: 8,
        sources: { osm: { type: 'raster', tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'], tileSize: 256, attribution: '© OpenStreetMap' } },
        layers: [{ id: 'osm', type: 'raster', source: 'osm', minzoom: 0, maxzoom: 19 }],
      },
      center: [35.55, 33.88], zoom: 12,
    });
    mapRef.current.addControl(new maplibregl.NavigationControl(), 'top-right');
    mapRef.current.addControl(new maplibregl.ScaleControl({ maxWidth: 200 }), 'bottom-left');
    return () => { markersRef.current.forEach(m => m.remove()); mapRef.current?.remove(); mapRef.current = null; };
  }, []);

  // ── Render layers ──
  useEffect(() => {
    const m = mapRef.current;
    if (!m) return;
    const render = () => {
      ['links-layer'].forEach(id => { if (m.getLayer(id)) m.removeLayer(id); });
      if (m.getSource('links-src')) m.removeSource('links-src');
      markersRef.current.forEach(mk => mk.remove());
      markersRef.current = [];

      if (layers.links) {
        const feats = sim.links.map(link => {
          const src = sim.routers.find(r => r.id === link.source_id);
          const tgt = sim.routers.find(r => r.id === link.target_id);
          if (!src || !tgt) return null;
          return { type: 'Feature' as const, geometry: { type: 'LineString' as const, coordinates: [src.coordinates, tgt.coordinates] }, properties: { color: linkColor(link), width: linkWidth(link.utilization) } };
        }).filter(Boolean);
        m.addSource('links-src', { type: 'geojson', data: { type: 'FeatureCollection', features: feats as GeoJSON.Feature[] } });
        m.addLayer({ id: 'links-layer', type: 'line', source: 'links-src', paint: { 'line-color': ['get', 'color'], 'line-width': ['get', 'width'], 'line-opacity': 0.85 }, layout: { 'line-cap': 'round', 'line-join': 'round' } });
      }

      if (layers.routers) {
        sim.routers.forEach(router => {
          const sz = nodeSize(router.role);
          const color = statusColors[router.status] || '#6B7280';
          const hasAlerts = router.active_alerts.length > 0;
          const hasCrit = router.active_alerts.some(a => a.severity === 'critical');

          const el = document.createElement('div');
          el.style.cssText = `position:relative;width:${sz + 16}px;height:${sz + 16}px;cursor:pointer;`;

          if (hasAlerts) {
            const pulse = document.createElement('div');
            pulse.style.cssText = `position:absolute;inset:0;border-radius:50%;border:2px solid ${hasCrit ? '#EF4444' : '#F59E0B'};animation:marker-pulse 2s ease-out infinite;opacity:0;`;
            el.appendChild(pulse);
          }

          const node = document.createElement('div');
          node.style.cssText = `position:absolute;left:50%;top:50%;transform:translate(-50%,-50%);width:${sz}px;height:${sz}px;background:${color};border:3px solid white;border-radius:50%;box-shadow:0 2px 8px rgba(0,0,0,0.3);transition:transform 0.2s,box-shadow 0.2s;display:flex;align-items:center;justify-content:center;`;

          const icon = document.createElement('div');
          const isz = sz * 0.45;
          if (router.role === 'core') icon.innerHTML = `<svg width="${isz}" height="${isz}" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2"><rect x="2" y="2" width="20" height="20" rx="2"/><line x1="12" y1="2" x2="12" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/></svg>`;
          else if (router.role === 'pppoe') icon.innerHTML = `<svg width="${isz}" height="${isz}" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/></svg>`;
          else icon.innerHTML = `<svg width="${isz}" height="${isz}" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2"><rect x="2" y="6" width="20" height="12" rx="2"/><circle cx="6" cy="12" r="1"/><circle cx="10" cy="12" r="1"/></svg>`;
          node.appendChild(icon);
          el.appendChild(node);

          if (hasAlerts) {
            const badge = document.createElement('div');
            badge.style.cssText = `position:absolute;top:0;right:0;background:${hasCrit ? '#EF4444' : '#F59E0B'};color:white;font-size:10px;font-weight:700;width:16px;height:16px;border-radius:50%;display:flex;align-items:center;justify-content:center;border:2px solid white;box-shadow:0 1px 3px rgba(0,0,0,0.3);`;
            badge.textContent = String(router.active_alerts.length);
            el.appendChild(badge);
          }

          if (router.pppoe_sessions > 0) {
            const lbl = document.createElement('div');
            lbl.style.cssText = `position:absolute;bottom:-18px;left:50%;transform:translateX(-50%);background:rgba(0,0,0,0.75);color:white;font-size:10px;font-weight:600;padding:1px 6px;border-radius:8px;white-space:nowrap;`;
            lbl.textContent = `${router.pppoe_sessions} users`;
            el.appendChild(lbl);
          }

          el.onmouseenter = () => { node.style.transform = 'translate(-50%,-50%) scale(1.15)'; node.style.boxShadow = '0 4px 12px rgba(0,0,0,0.4)'; };
          el.onmouseleave = () => { node.style.transform = 'translate(-50%,-50%) scale(1)'; node.style.boxShadow = '0 2px 8px rgba(0,0,0,0.3)'; };

          const popup = new maplibregl.Popup({ offset: 25, closeButton: false, maxWidth: '320px' }).setHTML(`
            <div style="padding:12px;font-family:system-ui,sans-serif;">
              <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px;"><div style="width:10px;height:10px;border-radius:50%;background:${color};"></div><strong>${router.name}</strong><span style="font-size:11px;padding:2px 8px;border-radius:12px;background:${color}15;color:${color};font-weight:600;margin-left:auto;">${router.status}</span></div>
              <div style="font-size:12px;color:#6B7280;margin-bottom:8px;">${router.vendor} ${router.model} · ${router.management_ip}</div>
              <div style="display:grid;grid-template-columns:1fr 1fr;gap:6px;font-size:12px;">
                <div style="background:#f9fafb;padding:6px 8px;border-radius:6px;"><div style="color:#9CA3AF;font-size:10px;">CPU</div><div style="font-weight:600;color:${router.cpu_usage > 80 ? '#EF4444' : router.cpu_usage > 60 ? '#F59E0B' : '#10B981'}">${router.cpu_usage.toFixed(1)}%</div></div>
                <div style="background:#f9fafb;padding:6px 8px;border-radius:6px;"><div style="color:#9CA3AF;font-size:10px;">Memory</div><div style="font-weight:600;color:${router.memory_usage > 85 ? '#EF4444' : router.memory_usage > 70 ? '#F59E0B' : '#10B981'}">${router.memory_usage.toFixed(1)}%</div></div>
                <div style="background:#f9fafb;padding:6px 8px;border-radius:6px;"><div style="color:#9CA3AF;font-size:10px;">Sessions</div><div style="font-weight:600;">${router.pppoe_sessions}${router.max_sessions ? '/' + router.max_sessions : ''}</div></div>
                <div style="background:#f9fafb;padding:6px 8px;border-radius:6px;"><div style="color:#9CA3AF;font-size:10px;">Uptime</div><div style="font-weight:600;">${formatUptime(router.uptime_seconds)}</div></div>
              </div>
              ${router.active_alerts.length > 0 ? `<div style="margin-top:8px;padding:6px 8px;background:#FEF2F240;border:1px solid #FECACA60;border-radius:6px;"><div style="font-size:11px;font-weight:600;color:#DC2626;">⚠ ${router.active_alerts.length} active alert${router.active_alerts.length > 1 ? 's' : ''}</div><div style="font-size:11px;color:#9CA3AF;margin-top:2px;">${router.active_alerts[0].title}</div></div>` : ''}
            </div>`);

          const marker = new maplibregl.Marker({ element: el }).setLngLat(router.coordinates).setPopup(popup).addTo(m);
          el.addEventListener('click', () => setSelectedNode(router));
          markersRef.current.push(marker);
        });
      }
    };
    if (m.isStyleLoaded()) render(); else m.on('load', render);
  }, [sim, layers]);

  const toggle = (k: keyof typeof layers) => setLayers(p => ({ ...p, [k]: !p[k] }));

  // ─── Render ───────────────────────────────────────────────────────────

  return (
    <div className="flex flex-col gap-4 h-full">
      {/* Header */}
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="text-xl font-bold text-gray-900 dark:text-white mr-auto">Network Map</h1>
        <Pill icon={<ServerStackIcon className="h-4 w-4 text-gray-500" />} label="nodes" value={sim.routers.length} extra={downCount > 0 ? <span className="text-red-600 dark:text-red-400 font-semibold">({downCount} down)</span> : null} />
        <Pill icon={<UserGroupIcon className="h-4 w-4 text-gray-500" />} label="PPPoE" value={totalUsers} />
        <Pill icon={<SignalIcon className="h-4 w-4 text-gray-500" />} label="" value={null} extra={<span className="text-gray-500 dark:text-gray-400 text-xs">↓{formatBps(bwIn)} ↑{formatBps(bwOut)}</span>} />
        {critAlerts > 0 && (
          <div className="flex items-center gap-1.5 rounded-lg bg-red-100 dark:bg-red-900/30 px-3 py-1.5 text-sm">
            <BellAlertIcon className="h-4 w-4 text-red-600 dark:text-red-400" />
            <span className="font-semibold text-red-700 dark:text-red-400">{critAlerts} critical</span>
          </div>
        )}
        <Button variant="secondary" size="sm" onClick={refresh} disabled={spinning}>
          <ArrowPathIcon className={cn('h-4 w-4 mr-1', spinning && 'animate-spin')} />Refresh
        </Button>
      </div>

      {/* Layer controls */}
      <Card className="p-2.5">
        <div className="flex flex-wrap items-center gap-3">
          <span className="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Layers</span>
          {(['routers', 'links', 'users'] as const).map(k => (
            <button key={k} onClick={() => toggle(k)} className={cn('flex items-center gap-1.5 rounded-md px-3 py-1 text-xs font-medium transition-colors', layers[k] ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400')}>
              {layers[k] ? <EyeIcon className="h-3.5 w-3.5" /> : <EyeSlashIcon className="h-3.5 w-3.5" />}
              {k.charAt(0).toUpperCase() + k.slice(1)}
            </button>
          ))}
          <div className="ml-auto flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
            <span className="font-medium">Status:</span>
            {[['Active', 'bg-green-500'], ['Degraded', 'bg-amber-500'], ['Offline', 'bg-red-500']].map(([l, c]) => (
              <div key={l} className="flex items-center gap-1"><div className={cn('h-2.5 w-2.5 rounded-full', c)} /><span>{l}</span></div>
            ))}
          </div>
        </div>
      </Card>

      {/* Map + panels */}
      <div className="flex gap-4 flex-1 min-h-0">
        <div className="flex-1 relative">
          <div ref={mapContainer} className="h-[600px] w-full rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm" />

          {selectedNode && (
            <div className="absolute left-4 top-4 w-80 z-10">
              <Card className="shadow-lg border border-gray-200 dark:border-gray-700">
                <div className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                      <div className="h-3 w-3 rounded-full" style={{ background: statusColors[selectedNode.status] }} />
                      <h3 className="font-bold text-gray-900 dark:text-white text-sm">{selectedNode.name}</h3>
                    </div>
                    <button onClick={() => setSelectedNode(null)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"><XMarkIcon className="h-4 w-4" /></button>
                  </div>
                  <div className="mt-1 text-xs text-gray-500 dark:text-gray-400">{selectedNode.vendor} {selectedNode.model} · {selectedNode.management_ip}</div>
                  <div className="mt-3 flex gap-1.5">
                    <Badge variant={selectedNode.status === 'active' ? 'success' : selectedNode.status === 'degraded' ? 'warning' : 'danger'}>{selectedNode.status}</Badge>
                    <Badge variant="default">{selectedNode.role}</Badge>
                  </div>
                  <div className="mt-3 grid grid-cols-2 gap-2">
                    <MBox label="CPU" value={`${selectedNode.cpu_usage.toFixed(1)}%`} lv={selectedNode.cpu_usage > 80 ? 'd' : selectedNode.cpu_usage > 60 ? 'w' : 'o'} pct={selectedNode.cpu_usage} />
                    <MBox label="Memory" value={`${selectedNode.memory_usage.toFixed(1)}%`} lv={selectedNode.memory_usage > 85 ? 'd' : selectedNode.memory_usage > 70 ? 'w' : 'o'} pct={selectedNode.memory_usage} />
                    <MBox label="Temp" value={selectedNode.temperature ? `${selectedNode.temperature.toFixed(0)}°C` : 'N/A'} lv={selectedNode.temperature > 60 ? 'd' : selectedNode.temperature > 50 ? 'w' : 'o'} pct={selectedNode.temperature ? (selectedNode.temperature / 80) * 100 : 0} />
                    <MBox label="Sessions" value={selectedNode.max_sessions ? `${selectedNode.pppoe_sessions}/${selectedNode.max_sessions}` : String(selectedNode.pppoe_sessions)} lv={selectedNode.max_sessions && selectedNode.pppoe_sessions / selectedNode.max_sessions > 0.85 ? 'd' : 'o'} pct={selectedNode.max_sessions ? (selectedNode.pppoe_sessions / selectedNode.max_sessions) * 100 : 0} />
                  </div>
                  <div className="mt-3 space-y-1 text-xs">
                    <div className="flex justify-between text-gray-500 dark:text-gray-400"><span>BW In</span><span className="font-medium text-gray-900 dark:text-white">{formatBps(selectedNode.bandwidth_in)}</span></div>
                    <div className="flex justify-between text-gray-500 dark:text-gray-400"><span>BW Out</span><span className="font-medium text-gray-900 dark:text-white">{formatBps(selectedNode.bandwidth_out)}</span></div>
                    <div className="flex justify-between text-gray-500 dark:text-gray-400"><span>Uptime</span><span className="font-medium text-gray-900 dark:text-white">{formatUptime(selectedNode.uptime_seconds)}</span></div>
                  </div>
                  {selectedNode.active_alerts.length > 0 && (
                    <div className="mt-3 space-y-1.5">
                      <div className="text-xs font-semibold text-gray-700 dark:text-gray-300">Active Alerts ({selectedNode.active_alerts.length})</div>
                      {selectedNode.active_alerts.map(a => (
                        <div key={a.id} className={cn('p-2 rounded border text-xs', sevBg(a.severity))}><div className={cn('font-semibold', sevColor(a.severity))}>{a.title}</div><div className="text-gray-600 dark:text-gray-400 mt-0.5">{a.affected_users} users affected</div></div>
                      ))}
                    </div>
                  )}
                  <div className="mt-3 rounded bg-gray-50 dark:bg-gray-800 p-2 text-xs text-gray-500 dark:text-gray-400">
                    <div className="font-semibold text-gray-700 dark:text-gray-300 mb-1">MikroTik API</div>
                    <div className="space-y-0.5">
                      <div>Endpoint: <code className="text-gray-900 dark:text-gray-200">{selectedNode.management_ip}:8728</code></div>
                      <div>Protocol: <code className="text-gray-900 dark:text-gray-200">RouterOS API + SNMP</code></div>
                      <div>Polls: <code className="text-gray-900 dark:text-gray-200">30s interval</code></div>
                    </div>
                  </div>
                </div>
              </Card>
            </div>
          )}
        </div>

        {/* Side panels */}
        <div className="w-80 flex flex-col gap-3 overflow-y-auto max-h-[600px]">
          {/* Alerts */}
          <Card>
            <button onClick={() => setAlertsOpen(!alertsOpen)} className="w-full flex items-center justify-between p-3 text-left">
              <div className="flex items-center gap-2">
                <BellAlertIcon className={cn('h-5 w-5', critAlerts > 0 ? 'text-red-500' : 'text-amber-500')} />
                <span className="font-semibold text-sm text-gray-900 dark:text-white">Proactive Alerts</span>
                {sim.alerts.length > 0 && <span className={cn('text-xs font-bold px-1.5 py-0.5 rounded-full', critAlerts > 0 ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400')}>{sim.alerts.length}</span>}
              </div>
              {alertsOpen ? <ChevronUpIcon className="h-4 w-4 text-gray-400" /> : <ChevronDownIcon className="h-4 w-4 text-gray-400" />}
            </button>
            {alertsOpen && (
              <div className="px-3 pb-3 space-y-2">
                {sim.alerts.length === 0 ? (
                  <div className="text-center text-xs text-gray-400 py-4"><CheckCircleIcon className="h-8 w-8 mx-auto mb-2 text-green-400" />All systems operational</div>
                ) : sim.alerts.sort((a, b) => ({ critical: 0, warning: 1, info: 2 }[a.severity] ?? 3) - ({ critical: 0, warning: 1, info: 2 }[b.severity] ?? 3)).map(alert => {
                  const router = sim.routers.find(r => r.id === alert.router_id);
                  const mins = Math.round((Date.now() - alert.detected_at) / 60_000);
                  return (
                    <div key={alert.id} className={cn('rounded-lg border p-2.5 text-xs', sevBg(alert.severity))}>
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex items-center gap-1.5">
                          <ExclamationTriangleIcon className={cn('h-4 w-4 shrink-0', alert.severity === 'critical' ? 'text-red-500' : alert.severity === 'warning' ? 'text-amber-500' : 'text-blue-500')} />
                          <span className={cn('font-bold', sevColor(alert.severity))}>{alert.title}</span>
                        </div>
                        {alert.auto_detected && <span className="shrink-0 rounded bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400 px-1.5 py-0.5 text-[10px] font-semibold">AUTO</span>}
                      </div>
                      <p className="mt-1 text-gray-600 dark:text-gray-400 leading-relaxed">{alert.description}</p>
                      <div className="mt-1.5 flex items-center gap-3 text-gray-500 dark:text-gray-400">
                        <span className="flex items-center gap-1"><UserGroupIcon className="h-3 w-3" />{alert.affected_users} affected</span>
                        <span className="flex items-center gap-1"><ClockIcon className="h-3 w-3" />{mins}m ago</span>
                      </div>
                      {router && <button onClick={() => setSelectedNode(router)} className="mt-1.5 text-blue-600 dark:text-blue-400 hover:underline font-medium">→ Locate: {router.name}</button>}
                      <div className="mt-1.5 p-1.5 rounded bg-white/60 dark:bg-gray-900/40 text-gray-600 dark:text-gray-400"><span className="font-semibold">Impact:</span> {alert.impact}</div>
                    </div>
                  );
                })}
              </div>
            )}
          </Card>

          {/* PPPoE sessions */}
          <Card>
            <button onClick={() => setUsersOpen(!usersOpen)} className="w-full flex items-center justify-between p-3 text-left">
              <div className="flex items-center gap-2">
                <WifiIcon className="h-5 w-5 text-blue-500" />
                <span className="font-semibold text-sm text-gray-900 dark:text-white">PPPoE Sessions</span>
                <span className="text-xs font-bold px-1.5 py-0.5 rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{totalUsers}</span>
              </div>
              {usersOpen ? <ChevronUpIcon className="h-4 w-4 text-gray-400" /> : <ChevronDownIcon className="h-4 w-4 text-gray-400" />}
            </button>
            {usersOpen && (
              <div className="px-3 pb-3 space-y-2">
                {sim.routers.filter(r => r.role === 'pppoe').map(router => {
                  const ru = sim.users.filter(u => u.router_id === router.id);
                  const rx = ru.reduce((s, u) => s + u.rx_rate, 0);
                  const tx = ru.reduce((s, u) => s + u.tx_rate, 0);
                  return (
                    <div key={router.id} onClick={() => setSelectedNode(router)} className={cn('p-2.5 rounded-lg border text-xs cursor-pointer transition-colors hover:bg-gray-50 dark:hover:bg-gray-800', router.status === 'offline' ? 'border-red-200 dark:border-red-800 bg-red-50/50 dark:bg-red-950/20' : 'border-gray-200 dark:border-gray-700')}>
                      <div className="flex items-center justify-between mb-1.5">
                        <div className="flex items-center gap-1.5"><div className="h-2 w-2 rounded-full" style={{ background: statusColors[router.status] }} /><span className="font-semibold text-gray-900 dark:text-white">{router.name}</span></div>
                        <span className="text-gray-500 dark:text-gray-400">{router.pppoe_sessions}/{router.max_sessions}</span>
                      </div>
                      <div className="h-1.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden mb-1.5">
                        <div className={cn('h-full rounded-full', router.max_sessions && router.pppoe_sessions / router.max_sessions > 0.85 ? 'bg-red-500' : router.max_sessions && router.pppoe_sessions / router.max_sessions > 0.7 ? 'bg-amber-500' : 'bg-green-500')} style={{ width: `${router.max_sessions ? (router.pppoe_sessions / router.max_sessions) * 100 : 0}%` }} />
                      </div>
                      <div className="flex justify-between text-gray-500 dark:text-gray-400"><span>↓{formatBps(rx)} ↑{formatBps(tx)}</span></div>
                      {router.status === 'offline' && <div className="mt-1.5 flex items-center gap-1 text-red-600 dark:text-red-400 font-medium"><XCircleIcon className="h-3.5 w-3.5" />Unreachable — sessions lost</div>}
                    </div>
                  );
                })}
                <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                  <div className="font-semibold text-gray-700 dark:text-gray-300 mb-1">Recent Sessions</div>
                  <div className="space-y-1 max-h-40 overflow-y-auto">
                    {sim.users.slice(0, 8).map(u => (
                      <div key={u.id} className="flex items-center justify-between p-1.5 rounded hover:bg-gray-50 dark:hover:bg-gray-800">
                        <div><div className="font-medium text-gray-900 dark:text-gray-200">{u.username.split('@')[0]}</div><div className="text-gray-400">{u.ip_address} · {u.plan}</div></div>
                        <div className="text-right"><div>↓{formatBps(u.rx_rate)}</div><div>↑{formatBps(u.tx_rate)}</div></div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </Card>

          {/* MikroTik info */}
          <Card>
            <div className="p-3">
              <div className="flex items-center gap-2 mb-2"><CpuChipIcon className="h-5 w-5 text-indigo-500" /><span className="font-semibold text-sm text-gray-900 dark:text-white">MikroTik Integration</span></div>
              <div className="space-y-1.5 text-xs text-gray-600 dark:text-gray-400">
                {['RouterOS API (8728) — sessions, interfaces', 'SNMP v2c/v3 — traffic, CPU, memory, temp', 'PPPoE monitoring — real-time user tracking', 'Proactive alerts — capacity, link down, thresholds'].map((t, i) => (
                  <div key={i} className="flex items-center gap-2"><CheckCircleIcon className="h-3.5 w-3.5 text-green-500 shrink-0" /><span>{t}</span></div>
                ))}
                <div className="flex items-center gap-2"><ClockIcon className="h-3.5 w-3.5 text-gray-400 shrink-0" /><span>Polling: 30s (configurable)</span></div>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}

// ─── Small sub-components ───────────────────────────────────────────────────

function Pill({ icon, label, value, extra }: { icon: React.ReactNode; label: string; value: number | null; extra?: React.ReactNode }) {
  return (
    <div className="flex items-center gap-1.5 rounded-lg bg-gray-100 dark:bg-gray-800 px-3 py-1.5 text-sm">
      {icon}
      {value !== null && <span className="font-semibold text-gray-900 dark:text-white">{value}</span>}
      {label && <span className="text-gray-500 dark:text-gray-400">{label}</span>}
      {extra}
    </div>
  );
}

function MBox({ label, value, lv, pct }: { label: string; value: string; lv: 'd' | 'w' | 'o'; pct: number }) {
  const bar = lv === 'd' ? 'bg-red-500' : lv === 'w' ? 'bg-amber-500' : 'bg-green-500';
  const txt = lv === 'd' ? 'text-red-600 dark:text-red-400' : lv === 'w' ? 'text-amber-600 dark:text-amber-400' : 'text-gray-900 dark:text-white';
  return (
    <div className="rounded-lg bg-gray-50 dark:bg-gray-800 p-2">
      <div className="text-[10px] font-medium text-gray-400 uppercase tracking-wider">{label}</div>
      <div className={cn('text-sm font-bold', txt)}>{value}</div>
      <div className="h-1 mt-1 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div className={cn('h-full rounded-full', bar)} style={{ width: `${Math.min(pct, 100)}%` }} />
      </div>
    </div>
  );
}

export default MapPage;

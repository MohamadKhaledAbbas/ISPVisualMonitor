import { useEffect, useRef, useState } from 'react';
import maplibregl, { Map, Marker } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { useQuery } from '@tanstack/react-query';
import { Card, Badge, Button } from '@/components/common';
import { topologyApi } from '@/api';
import { cn } from '@/utils';
import type { TopologyGeoJSON } from '@/types';
import {
  EyeIcon,
  EyeSlashIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';

// Mock topology data for development
const mockTopologyGeoJSON: TopologyGeoJSON = {
  type: 'FeatureCollection',
  features: [
    {
      type: 'Feature',
      geometry: { type: 'Point', coordinates: [-74.006, 40.7128] },
      properties: { type: 'router', id: '1', name: 'NYC-Core-1', status: 'active' },
    },
    {
      type: 'Feature',
      geometry: { type: 'Point', coordinates: [-118.2437, 34.0522] },
      properties: { type: 'router', id: '2', name: 'LA-Core-1', status: 'active' },
    },
    {
      type: 'Feature',
      geometry: { type: 'Point', coordinates: [-87.6298, 41.8781] },
      properties: { type: 'router', id: '3', name: 'Chicago-Core-1', status: 'offline' },
    },
    {
      type: 'Feature',
      geometry: { type: 'Point', coordinates: [-95.3698, 29.7604] },
      properties: { type: 'router', id: '4', name: 'Houston-Edge-1', status: 'active' },
    },
    {
      type: 'Feature',
      geometry: { type: 'Point', coordinates: [-122.4194, 37.7749] },
      properties: { type: 'router', id: '5', name: 'SF-Edge-1', status: 'maintenance' },
    },
    // Links
    {
      type: 'Feature',
      geometry: {
        type: 'LineString',
        coordinates: [
          [-74.006, 40.7128],
          [-87.6298, 41.8781],
        ],
      },
      properties: { type: 'link', id: 'link-1-3', status: 'active' },
    },
    {
      type: 'Feature',
      geometry: {
        type: 'LineString',
        coordinates: [
          [-87.6298, 41.8781],
          [-118.2437, 34.0522],
        ],
      },
      properties: { type: 'link', id: 'link-3-2', status: 'degraded' },
    },
    {
      type: 'Feature',
      geometry: {
        type: 'LineString',
        coordinates: [
          [-118.2437, 34.0522],
          [-122.4194, 37.7749],
        ],
      },
      properties: { type: 'link', id: 'link-2-5', status: 'active' },
    },
    {
      type: 'Feature',
      geometry: {
        type: 'LineString',
        coordinates: [
          [-87.6298, 41.8781],
          [-95.3698, 29.7604],
        ],
      },
      properties: { type: 'link', id: 'link-3-4', status: 'active' },
    },
  ],
};

const statusColors: Record<string, string> = {
  active: '#10B981',
  online: '#10B981',
  offline: '#EF4444',
  down: '#EF4444',
  maintenance: '#F59E0B',
  degraded: '#F59E0B',
};

export function MapPage() {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<Map | null>(null);
  const markersRef = useRef<Marker[]>([]);
  const [selectedRouter, setSelectedRouter] = useState<typeof mockTopologyGeoJSON.features[0]['properties'] | null>(null);
  const [layers, setLayers] = useState({
    routers: true,
    links: true,
    labels: true,
  });

  // In production, fetch from API
  const { data: topology, refetch } = useQuery({
    queryKey: ['topology-geojson'],
    queryFn: () => topologyApi.getGeoJSON(),
    enabled: false, // Using mock data
  });

  const geoData = topology || mockTopologyGeoJSON;

  useEffect(() => {
    if (!mapContainer.current || map.current) return;

    // Initialize map
    map.current = new maplibregl.Map({
      container: mapContainer.current,
      style: {
        version: 8,
        sources: {
          'osm': {
            type: 'raster',
            tiles: ['https://tile.openstreetmap.org/{z}/{x}/{y}.png'],
            tileSize: 256,
            attribution: '© OpenStreetMap contributors',
          },
        },
        layers: [
          {
            id: 'osm-tiles',
            type: 'raster',
            source: 'osm',
            minzoom: 0,
            maxzoom: 19,
          },
        ],
      },
      center: [-98.5795, 39.8283], // Center of US
      zoom: 3,
    });

    map.current.addControl(new maplibregl.NavigationControl(), 'top-right');

    return () => {
      markersRef.current.forEach((marker) => marker.remove());
      map.current?.remove();
      map.current = null;
    };
  }, []);

  // Add topology data to map
  useEffect(() => {
    if (!map.current || !geoData) return;

    const mapInstance = map.current;

    // Wait for map to load
    const addLayers = () => {
      // Remove existing layers and sources
      if (mapInstance.getLayer('links')) {
        mapInstance.removeLayer('links');
      }
      if (mapInstance.getSource('topology-links')) {
        mapInstance.removeSource('topology-links');
      }

      // Clear existing markers
      markersRef.current.forEach((marker) => marker.remove());
      markersRef.current = [];

      // Add links as a line layer
      const linkFeatures = geoData.features.filter((f) => f.properties.type === 'link');
      if (linkFeatures.length > 0 && layers.links) {
        mapInstance.addSource('topology-links', {
          type: 'geojson',
          data: {
            type: 'FeatureCollection',
            features: linkFeatures as GeoJSON.Feature[],
          },
        });

        mapInstance.addLayer({
          id: 'links',
          type: 'line',
          source: 'topology-links',
          paint: {
            'line-color': ['match', ['get', 'status'],
              'active', '#10B981',
              'degraded', '#F59E0B',
              '#EF4444'
            ],
            'line-width': 3,
            'line-opacity': 0.8,
          },
        });
      }

      // Add router markers
      if (layers.routers) {
        const routerFeatures = geoData.features.filter((f) => f.properties.type === 'router');
        routerFeatures.forEach((feature) => {
          if (feature.geometry.type !== 'Point') return;

          const coords = feature.geometry.coordinates as [number, number];
          const props = feature.properties;
          const color = statusColors[props.status || 'active'] || '#6B7280';

          // Create custom marker element
          const el = document.createElement('div');
          el.className = 'router-marker';
          el.style.cssText = `
            width: 24px;
            height: 24px;
            background-color: ${color};
            border: 3px solid white;
            border-radius: 50%;
            cursor: pointer;
            box-shadow: 0 2px 4px rgba(0,0,0,0.3);
            transition: transform 0.2s;
          `;

          el.addEventListener('mouseenter', () => {
            el.style.transform = 'scale(1.2)';
          });
          el.addEventListener('mouseleave', () => {
            el.style.transform = 'scale(1)';
          });

          // Create popup
          const popup = new maplibregl.Popup({
            offset: 25,
            closeButton: false,
          }).setHTML(`
            <div style="padding: 8px;">
              <h3 style="font-weight: 600; margin-bottom: 4px;">${props.name || 'Unknown'}</h3>
              <p style="color: #6B7280; font-size: 14px; margin-bottom: 4px;">ID: ${props.id}</p>
              <span style="
                display: inline-block;
                padding: 2px 8px;
                border-radius: 9999px;
                font-size: 12px;
                font-weight: 500;
                background-color: ${color}20;
                color: ${color};
              ">${props.status || 'unknown'}</span>
            </div>
          `);

          const marker = new maplibregl.Marker({ element: el })
            .setLngLat(coords)
            .setPopup(popup)
            .addTo(mapInstance);

          el.addEventListener('click', () => {
            setSelectedRouter(props);
          });

          markersRef.current.push(marker);
        });
      }
    };

    if (mapInstance.isStyleLoaded()) {
      addLayers();
    } else {
      mapInstance.on('load', addLayers);
    }
  }, [geoData, layers]);

  const toggleLayer = (layer: keyof typeof layers) => {
    setLayers((prev) => ({ ...prev, [layer]: !prev[layer] }));
  };

  return (
    <div className="space-y-4">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Network Map</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Interactive topology visualization
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => refetch()}
          >
            <ArrowPathIcon className="h-4 w-4 mr-1" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Map controls */}
      <Card className="p-3">
        <div className="flex flex-wrap items-center gap-4">
          <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
            Layers:
          </span>
          <button
            onClick={() => toggleLayer('routers')}
            className={cn(
              'flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              layers.routers
                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
            )}
          >
            {layers.routers ? (
              <EyeIcon className="h-4 w-4" />
            ) : (
              <EyeSlashIcon className="h-4 w-4" />
            )}
            Routers
          </button>
          <button
            onClick={() => toggleLayer('links')}
            className={cn(
              'flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              layers.links
                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
            )}
          >
            {layers.links ? (
              <EyeIcon className="h-4 w-4" />
            ) : (
              <EyeSlashIcon className="h-4 w-4" />
            )}
            Links
          </button>
          
          {/* Legend */}
          <div className="ml-auto flex items-center gap-4">
            <span className="text-sm text-gray-500 dark:text-gray-400">Status:</span>
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-1">
                <div className="h-3 w-3 rounded-full bg-green-500" />
                <span className="text-xs text-gray-600 dark:text-gray-400">Active</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="h-3 w-3 rounded-full bg-yellow-500" />
                <span className="text-xs text-gray-600 dark:text-gray-400">Maintenance</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="h-3 w-3 rounded-full bg-red-500" />
                <span className="text-xs text-gray-600 dark:text-gray-400">Offline</span>
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* Map container */}
      <div className="relative">
        <div
          ref={mapContainer}
          className="h-[600px] w-full rounded-lg border border-gray-200 dark:border-gray-700"
        />

        {/* Router details panel */}
        {selectedRouter && (
          <div className="absolute right-4 top-4 w-80">
            <Card>
              <div className="p-4">
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-semibold text-gray-900 dark:text-white">
                      {selectedRouter.name}
                    </h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      ID: {selectedRouter.id}
                    </p>
                  </div>
                  <button
                    onClick={() => setSelectedRouter(null)}
                    className="text-gray-400 hover:text-gray-500"
                  >
                    ×
                  </button>
                </div>
                <div className="mt-4">
                  <Badge
                    variant={
                      selectedRouter.status === 'active'
                        ? 'success'
                        : selectedRouter.status === 'maintenance'
                        ? 'warning'
                        : 'danger'
                    }
                  >
                    {selectedRouter.status}
                  </Badge>
                </div>
                <div className="mt-4 flex gap-2">
                  <Button size="sm" className="flex-1">
                    View Details
                  </Button>
                  <Button size="sm" variant="secondary">
                    Metrics
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        )}
      </div>
    </div>
  );
}

export default MapPage;

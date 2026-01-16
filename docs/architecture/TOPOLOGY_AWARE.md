# Topology-Aware Architecture

## Overview

The ISP Visual Monitor platform represents network infrastructure as a geographic topology graph, enabling intuitive visualization and intelligent monitoring based on network relationships and physical locations.

## Topology Model

### Network Entity Types

#### 1. Routers
- **Core Routers**: High-capacity backbone routers
- **Edge Routers**: Customer-facing routers
- **Border Routers**: Inter-network gateway routers

#### 2. Network Links
- **Physical Links**: Fiber, wireless, copper connections
- **Logical Links**: VPN tunnels, MPLS paths
- **Aggregated Links**: Link bundles (LAG/LACP)

#### 3. Points of Presence (POPs)
- Geographic locations housing network equipment
- Represented as polygons or points on the map
- Contain multiple routers and devices

#### 4. Geographic Regions
- Hierarchical organization of POPs
- Administrative boundaries
- Service areas

### Topology Graph Structure

```
Region (e.g., "North America")
  └── POP (e.g., "NYC-DC1")
      ├── Router (e.g., "NYC-CORE-01")
      │   ├── Interface (e.g., "eth0")
      │   └── Interface (e.g., "eth1")
      └── Router (e.g., "NYC-EDGE-01")
          └── Interfaces...

Links connect Interfaces between Routers
```

## Database Schema for Topology

### Core Tables

```sql
-- Geographic Regions
CREATE TABLE regions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    parent_region_id UUID REFERENCES regions(id),
    boundary GEOGRAPHY(POLYGON),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Points of Presence
CREATE TABLE pops (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    region_id UUID REFERENCES regions(id),
    name VARCHAR(255) NOT NULL,
    location GEOGRAPHY(POINT) NOT NULL,
    address TEXT,
    pop_type VARCHAR(50), -- datacenter, co-location, edge
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Routers
CREATE TABLE routers (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    pop_id UUID REFERENCES pops(id),
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255),
    management_ip INET NOT NULL,
    location GEOGRAPHY(POINT),
    router_type VARCHAR(50), -- core, edge, border
    vendor VARCHAR(100),
    model VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_router_name UNIQUE(tenant_id, name)
);

-- Router Interfaces
CREATE TABLE interfaces (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    router_id UUID NOT NULL REFERENCES routers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    if_index INTEGER,
    if_type VARCHAR(50),
    speed_mbps BIGINT,
    status VARCHAR(50) DEFAULT 'up',
    ip_address INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_interface UNIQUE(router_id, name)
);

-- Network Links
CREATE TABLE links (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255),
    source_interface_id UUID NOT NULL REFERENCES interfaces(id),
    target_interface_id UUID NOT NULL REFERENCES interfaces(id),
    link_type VARCHAR(50), -- physical, logical, vpn
    capacity_mbps BIGINT,
    latency_ms DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'up',
    path_geometry GEOGRAPHY(LINESTRING),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_link UNIQUE(source_interface_id, target_interface_id)
);
```

## Geographic Awareness

### Coordinate System
- **Standard**: WGS84 (EPSG:4326)
- **Storage**: PostGIS GEOGRAPHY type
- **Precision**: Sufficient for infrastructure-level accuracy

### Location-Based Features

#### 1. Distance Calculations
```sql
-- Find routers within 50km of a point
SELECT r.name, ST_Distance(r.location, ST_MakePoint(-73.9857, 40.7484)::geography) / 1000 as distance_km
FROM routers r
WHERE ST_DWithin(r.location, ST_MakePoint(-73.9857, 40.7484)::geography, 50000)
ORDER BY distance_km;
```

#### 2. Regional Containment
```sql
-- Find all POPs within a region
SELECT p.*
FROM pops p
JOIN regions r ON ST_Contains(r.boundary, p.location);
```

#### 3. Link Path Visualization
- Store link geometry as LINESTRING
- Visualize physical fiber routes on map
- Calculate geographic distance vs network distance

## Topology-Aware Monitoring

### 1. Impact Analysis

When a router fails, automatically determine:
- **Directly Affected**: Devices connected to failed router
- **Indirectly Affected**: Devices downstream in topology
- **Alternative Paths**: Redundant routes if available

```go
func AnalyzeImpact(routerID uuid.UUID, topology *NetworkTopology) ImpactReport {
    affected := topology.GetDownstreamDevices(routerID)
    alternativePaths := topology.FindAlternativePaths(affected)
    
    return ImpactReport{
        DirectlyAffected: affected.Direct,
        IndirectlyAffected: affected.Indirect,
        HasRedundancy: alternativePaths != nil,
        CustomerImpact: calculateCustomerImpact(affected),
    }
}
```

### 2. Path-Based Alerts

Alert based on network paths:
- "Link utilization > 80% on primary path"
- "No redundant path available for critical circuit"
- "Latency increased on path between NYC and SF"

### 3. Geographic Correlation

Correlate alerts with geographic events:
- Multiple failures in same region
- Weather-based correlations
- Maintenance windows by location

## Topology Discovery

### Manual Configuration
- Web UI for adding devices and links
- Bulk import via CSV/JSON
- API for programmatic configuration

### Automatic Discovery
- LLDP/CDP neighbor discovery
- SNMP topology polling
- BGP peer relationships

```go
func DiscoverTopology(router *Router) []Link {
    // Query LLDP neighbors
    neighbors := snmp.GetLLDPNeighbors(router.ManagementIP)
    
    links := []Link{}
    for _, neighbor := range neighbors {
        // Find neighbor router in database
        neighborRouter := findRouterByHostname(neighbor.Hostname)
        
        if neighborRouter != nil {
            link := Link{
                SourceInterface: router.GetInterface(neighbor.LocalPort),
                TargetInterface: neighborRouter.GetInterface(neighbor.RemotePort),
                LinkType: "physical",
            }
            links = append(links, link)
        }
    }
    
    return links
}
```

## Map Visualization

### Self-Hosted Maps with MapLibre

#### PMTiles Setup
- **Tile Storage**: PMTiles format for efficient serving
- **Basemap**: OpenStreetMap data or custom basemap
- **Overlay**: Network topology layer

#### Topology Layer
```javascript
map.addSource('network-topology', {
    type: 'geojson',
    data: '/api/topology/geojson'
});

// Router points
map.addLayer({
    id: 'routers',
    type: 'circle',
    source: 'network-topology',
    filter: ['==', '$type', 'Point'],
    paint: {
        'circle-radius': 8,
        'circle-color': [
            'match',
            ['get', 'status'],
            'up', '#00ff00',
            'down', '#ff0000',
            'warning', '#ffaa00',
            '#cccccc'
        ]
    }
});

// Links
map.addLayer({
    id: 'links',
    type: 'line',
    source: 'network-topology',
    filter: ['==', '$type', 'LineString'],
    paint: {
        'line-color': '#0088ff',
        'line-width': 2
    }
});
```

### Interactive Features
- Click router for details panel
- Hover for quick stats
- Draw custom regions
- Filter by status/type
- Search and zoom to device

## Topology Queries

### Graph Traversal

```go
// Find shortest path between two routers
func FindShortestPath(source, target uuid.UUID, topology *NetworkTopology) []Router {
    // Dijkstra's algorithm on topology graph
    graph := topology.BuildGraph()
    path := graph.ShortestPath(source, target)
    return path
}

// Find all paths (redundancy check)
func FindAllPaths(source, target uuid.UUID, topology *NetworkTopology) [][]Router {
    graph := topology.BuildGraph()
    paths := graph.AllPaths(source, target, maxHops)
    return paths
}
```

### Hierarchical Queries

```sql
-- Recursive query for region hierarchy
WITH RECURSIVE region_tree AS (
    SELECT id, name, parent_region_id, 0 as level
    FROM regions
    WHERE parent_region_id IS NULL AND tenant_id = $1
    
    UNION ALL
    
    SELECT r.id, r.name, r.parent_region_id, rt.level + 1
    FROM regions r
    JOIN region_tree rt ON r.parent_region_id = rt.id
)
SELECT * FROM region_tree ORDER BY level, name;
```

## Performance Optimization

### Spatial Indexes
```sql
CREATE INDEX idx_routers_location ON routers USING GIST(location);
CREATE INDEX idx_pops_location ON pops USING GIST(location);
CREATE INDEX idx_regions_boundary ON regions USING GIST(boundary);
CREATE INDEX idx_links_geometry ON links USING GIST(path_geometry);
```

### Materialized Views
```sql
-- Pre-computed topology graph for quick access
CREATE MATERIALIZED VIEW topology_graph AS
SELECT 
    l.id as link_id,
    l.source_interface_id,
    l.target_interface_id,
    r1.id as source_router_id,
    r2.id as target_router_id,
    l.capacity_mbps,
    l.status
FROM links l
JOIN interfaces i1 ON l.source_interface_id = i1.id
JOIN interfaces i2 ON l.target_interface_id = i2.id
JOIN routers r1 ON i1.router_id = r1.id
JOIN routers r2 ON i2.router_id = r2.id;

REFRESH MATERIALIZED VIEW topology_graph;
```

## Future Enhancements

### Advanced Visualization
- 3D topology view for datacenter racks
- Time-lapse replay of topology changes
- Heatmaps for utilization and latency

### ML-Based Insights
- Anomaly detection in topology
- Predictive capacity planning
- Automated path optimization

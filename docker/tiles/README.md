# Map Tiles Directory

This directory is for storing PMTiles files for MapLibre GL JS.

## Setting up map tiles

### Option 1: Download pre-built PMTiles

You can download pre-built PMTiles from various sources:

- [OpenMapTiles](https://openmaptiles.org/)
- [Protomaps](https://protomaps.com/)
- Custom PMTiles generated from OpenStreetMap data

Place the `.pmtiles` files in this directory.

### Option 2: Generate your own PMTiles

Use tools like:
- [tippecanoe](https://github.com/felt/tippecanoe) - Convert GeoJSON to PMTiles
- [planetiler](https://github.com/onthegomap/planetiler) - Generate vector tiles from OSM data

### Usage in MapLibre

```javascript
import maplibregl from 'maplibre-gl';
import { Protocol } from 'pmtiles';

// Register PMTiles protocol
let protocol = new Protocol();
maplibregl.addProtocol("pmtiles", protocol.tile);

// Create map with PMTiles source
const map = new maplibregl.Map({
  container: 'map',
  style: {
    version: 8,
    sources: {
      'basemap': {
        type: 'vector',
        url: 'pmtiles:///tiles/basemap.pmtiles'
      }
    },
    layers: [
      // Define your layers here
    ]
  }
});
```

## Directory Structure

```
tiles/
├── basemap.pmtiles        # Base map tiles
├── satellite.pmtiles      # Satellite imagery (optional)
└── terrain.pmtiles        # Terrain data (optional)
```

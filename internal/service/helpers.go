package service

import (
	"fmt"
	"strings"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/api/dto"
)

// postGISToGeoPoint converts PostGIS POINT(lon lat) string to GeoPoint
func postGISToGeoPoint(location string) *dto.GeoPoint {
	// Parse "POINT(lon lat)" format
	location = strings.TrimPrefix(location, "POINT(")
	location = strings.TrimSuffix(location, ")")
	parts := strings.Fields(location)
	if len(parts) != 2 {
		return nil
	}

	var lon, lat float64
	if _, err := fmt.Sscanf(parts[0], "%f", &lon); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(parts[1], "%f", &lat); err != nil {
		return nil
	}

	return &dto.GeoPoint{
		Latitude:  lat,
		Longitude: lon,
	}
}

// geoPointToPostGIS converts GeoPoint to PostGIS POINT(lon lat) string
func geoPointToPostGIS(point *dto.GeoPoint) string {
	return fmt.Sprintf("POINT(%f %f)", point.Longitude, point.Latitude)
}

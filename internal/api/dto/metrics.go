package dto

import (
	"time"
)

// MetricDataPoint represents a single metric data point
type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// InterfaceMetrics represents metrics for a network interface
type InterfaceMetrics struct {
	InterfaceID       string            `json:"interface_id"`
	InterfaceName     string            `json:"interface_name"`
	InBps             []MetricDataPoint `json:"in_bps,omitempty"`
	OutBps            []MetricDataPoint `json:"out_bps,omitempty"`
	InPackets         []MetricDataPoint `json:"in_packets,omitempty"`
	OutPackets        []MetricDataPoint `json:"out_packets,omitempty"`
	InErrors          []MetricDataPoint `json:"in_errors,omitempty"`
	OutErrors         []MetricDataPoint `json:"out_errors,omitempty"`
	Utilization       []MetricDataPoint `json:"utilization,omitempty"`
}

// RouterMetrics represents metrics for a router
type RouterMetrics struct {
	RouterID        string            `json:"router_id"`
	RouterName      string            `json:"router_name"`
	CPUUsage        []MetricDataPoint `json:"cpu_usage,omitempty"`
	MemoryUsage     []MetricDataPoint `json:"memory_usage,omitempty"`
	TotalInBps      []MetricDataPoint `json:"total_in_bps,omitempty"`
	TotalOutBps     []MetricDataPoint `json:"total_out_bps,omitempty"`
	ActiveInterfaces []MetricDataPoint `json:"active_interfaces,omitempty"`
}

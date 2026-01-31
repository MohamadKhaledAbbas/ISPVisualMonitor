package models

import (
	"time"

	"github.com/google/uuid"
)

// RouterCapabilities represents all connection methods and credentials for a router
type RouterCapabilities struct {
	ID       uuid.UUID `json:"id" db:"id"`
	RouterID uuid.UUID `json:"router_id" db:"router_id"`
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`

	// SNMP Capability
	SNMP *SNMPCapability `json:"snmp,omitempty"`

	// API Capability (vendor-specific)
	API *APICapability `json:"api,omitempty"`

	// SSH Capability
	SSH *SSHCapability `json:"ssh,omitempty"`

	// NETCONF Capability
	NETCONF *NETCONFCapability `json:"netconf,omitempty"`

	// Syslog Capability (passive)
	Syslog *SyslogCapability `json:"syslog,omitempty"`

	// NetFlow Capability (passive)
	NetFlow *NetFlowCapability `json:"netflow,omitempty"`

	// IPFIX Capability (passive)
	IPFIX *IPFIXCapability `json:"ipfix,omitempty"`

	// Connection preferences
	PreferredMethod string   `json:"preferred_method" db:"preferred_method"` // api, snmp, ssh, netconf
	FallbackOrder   []string `json:"fallback_order" db:"fallback_order"`     // Order of methods to try

	// Testing metadata
	LastTestedAt    *time.Time `json:"last_tested_at,omitempty" db:"last_tested_at"`
	LastTestSuccess *bool      `json:"last_test_success,omitempty" db:"last_test_success"`
	Notes           *string    `json:"notes,omitempty" db:"notes"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SNMPCapability represents SNMP connection details
type SNMPCapability struct {
	Enabled        bool    `json:"enabled" db:"snmp_enabled"`
	Version        string  `json:"version" db:"snmp_version"`               // v1, v2c, v3
	Community      *string `json:"community,omitempty" db:"snmp_community"` // For v1/v2c (sensitive)
	Port           int     `json:"port" db:"snmp_port"`
	TimeoutSeconds int     `json:"timeout_seconds" db:"snmp_timeout_seconds"`
	Retries        int     `json:"retries" db:"snmp_retries"`

	// SNMPv3 specific
	V3Username     *string `json:"v3_username,omitempty" db:"snmp_v3_username"`
	V3AuthProtocol *string `json:"v3_auth_protocol,omitempty" db:"snmp_v3_auth_protocol"` // MD5, SHA, SHA-256, etc.
	V3AuthPassword *string `json:"v3_auth_password,omitempty" db:"snmp_v3_auth_password"` // (sensitive)
	V3PrivProtocol *string `json:"v3_priv_protocol,omitempty" db:"snmp_v3_priv_protocol"` // DES, AES, AES-256
	V3PrivPassword *string `json:"v3_priv_password,omitempty" db:"snmp_v3_priv_password"` // (sensitive)
}

// APICapability represents vendor-specific API connection details
type APICapability struct {
	Enabled        bool   `json:"enabled" db:"api_enabled"`
	Type           string `json:"type" db:"api_type"` // mikrotik, cisco_restconf, juniper_netconf, arista_eapi
	Endpoint       string `json:"endpoint" db:"api_endpoint"`
	Port           *int   `json:"port,omitempty" db:"api_port"`
	Username       string `json:"username" db:"api_username"`
	Password       string `json:"password" db:"api_password"` // (sensitive)
	UseTLS         bool   `json:"use_tls" db:"api_use_tls"`
	VerifyCert     bool   `json:"verify_cert" db:"api_verify_cert"`
	TimeoutSeconds int    `json:"timeout_seconds" db:"api_timeout_seconds"`
}

// SSHCapability represents SSH connection details
type SSHCapability struct {
	Enabled        bool    `json:"enabled" db:"ssh_enabled"`
	Host           string  `json:"host" db:"ssh_host"`
	Port           int     `json:"port" db:"ssh_port"`
	Username       string  `json:"username" db:"ssh_username"`
	Password       *string `json:"password,omitempty" db:"ssh_password"`       // (sensitive)
	PrivateKey     *string `json:"private_key,omitempty" db:"ssh_private_key"` // (sensitive)
	TimeoutSeconds int     `json:"timeout_seconds" db:"ssh_timeout_seconds"`
}

// NETCONFCapability represents NETCONF connection details
type NETCONFCapability struct {
	Enabled  bool   `json:"enabled" db:"netconf_enabled"`
	Port     int    `json:"port" db:"netconf_port"`
	Username string `json:"username" db:"netconf_username"`
	Password string `json:"password" db:"netconf_password"` // (sensitive)
}

// SyslogCapability represents syslog receiver configuration
type SyslogCapability struct {
	Enabled  bool    `json:"enabled" db:"syslog_enabled"`
	Facility *string `json:"facility,omitempty" db:"syslog_facility"`
	Severity *string `json:"severity,omitempty" db:"syslog_severity"`
}

// NetFlowCapability represents NetFlow/sFlow export configuration
type NetFlowCapability struct {
	Enabled bool `json:"enabled" db:"netflow_enabled"`
	Version *int `json:"version,omitempty" db:"netflow_version"` // 5, 9, 10 (IPFIX)
	Port    int  `json:"port" db:"netflow_port"`
}

// IPFIXCapability represents IPFIX export configuration
type IPFIXCapability struct {
	Enabled bool `json:"enabled" db:"ipfix_enabled"`
	Port    int  `json:"port" db:"ipfix_port"`
}

// GetEnabledMethods returns a list of enabled connection methods
func (rc *RouterCapabilities) GetEnabledMethods() []string {
	methods := []string{}

	if rc.SNMP != nil && rc.SNMP.Enabled {
		methods = append(methods, "snmp")
	}
	if rc.API != nil && rc.API.Enabled {
		methods = append(methods, "api")
	}
	if rc.SSH != nil && rc.SSH.Enabled {
		methods = append(methods, "ssh")
	}
	if rc.NETCONF != nil && rc.NETCONF.Enabled {
		methods = append(methods, "netconf")
	}
	if rc.Syslog != nil && rc.Syslog.Enabled {
		methods = append(methods, "syslog")
	}
	if rc.NetFlow != nil && rc.NetFlow.Enabled {
		methods = append(methods, "netflow")
	}
	if rc.IPFIX != nil && rc.IPFIX.Enabled {
		methods = append(methods, "ipfix")
	}

	return methods
}

// HasActivePollingMethod checks if router has at least one active polling method
func (rc *RouterCapabilities) HasActivePollingMethod() bool {
	if rc.SNMP != nil && rc.SNMP.Enabled {
		return true
	}
	if rc.API != nil && rc.API.Enabled {
		return true
	}
	if rc.SSH != nil && rc.SSH.Enabled {
		return true
	}
	if rc.NETCONF != nil && rc.NETCONF.Enabled {
		return true
	}
	return false
}

// GetPreferredMethodOrder returns the preferred order of methods to try
func (rc *RouterCapabilities) GetPreferredMethodOrder() []string {
	if len(rc.FallbackOrder) > 0 {
		return rc.FallbackOrder
	}

	// Default fallback order
	order := []string{}
	if rc.PreferredMethod != "" {
		order = append(order, rc.PreferredMethod)
	}

	// Add other enabled methods
	enabled := rc.GetEnabledMethods()
	for _, method := range enabled {
		if method != rc.PreferredMethod {
			order = append(order, method)
		}
	}

	return order
}

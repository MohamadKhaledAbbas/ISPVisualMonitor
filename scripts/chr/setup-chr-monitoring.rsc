# MikroTik RouterOS Configuration Script
# This script configures CHR instances for ISP Visual Monitor testing
# Run this script on each CHR instance to enable monitoring capabilities

# ============================================================================
# SYSTEM IDENTITY AND USER MANAGEMENT
# ============================================================================

/system identity set name="CHR-Test-Router"

# Create monitoring user
/user add name=monitor password=monitor123 group=read comment="ISP Monitor polling user"

# ============================================================================
# SNMP CONFIGURATION
# ============================================================================

# Enable SNMP
/snmp set enabled=yes contact="admin@example.com" location="CHR Lab" \
    trap-community=public trap-version=2

# Add SNMP community
/snmp community add name=public addresses=0.0.0.0/0 \
    read-access=yes write-access=no

# ============================================================================
# API CONFIGURATION
# ============================================================================

# Enable API (default port 8728)
/ip service enable api
/ip service set api address=0.0.0.0/0 port=8728

# Enable API-SSL (default port 8729)
/ip service enable api-ssl
/ip service set api-ssl address=0.0.0.0/0 port=8729 certificate=none

# Enable SSH
/ip service enable ssh
/ip service set ssh address=0.0.0.0/0 port=22

# Enable WebFig (web interface)
/ip service enable www
/ip service set www address=0.0.0.0/0 port=80

# ============================================================================
# PPPOE SERVER CONFIGURATION (for pppoe_server role)
# ============================================================================

# Create IP pool for PPPoE clients
/ip pool add name=pppoe-pool ranges=10.10.10.2-10.10.10.254

# Create PPPoE server profile
/ppp profile add name=pppoe-profile local-address=10.10.10.1 \
    remote-address=pppoe-pool dns-server=8.8.8.8,8.8.4.4 \
    use-compression=no use-encryption=no only-one=no

# Create PPPoE server
/interface pppoe-server server add \
    service-name=ISP-PPPoE \
    interface=ether1 \
    default-profile=pppoe-profile \
    authentication=pap,chap,mschap1,mschap2 \
    keepalive-timeout=60 \
    one-session-per-host=no

# Add test PPPoE users
/ppp secret add name=user001 password=pass001 service=pppoe profile=pppoe-profile comment="Test User 1"
/ppp secret add name=user002 password=pass002 service=pppoe profile=pppoe-profile comment="Test User 2"
/ppp secret add name=user003 password=pass003 service=pppoe profile=pppoe-profile comment="Test User 3"
/ppp secret add name=user004 password=pass004 service=pppoe profile=pppoe-profile comment="Test User 4"
/ppp secret add name=user005 password=pass005 service=pppoe profile=pppoe-profile comment="Test User 5"
/ppp secret add name=user006 password=pass006 service=pppoe profile=pppoe-profile comment="Test User 6"
/ppp secret add name=user007 password=pass007 service=pppoe profile=pppoe-profile comment="Test User 7"
/ppp secret add name=user008 password=pass008 service=pppoe profile=pppoe-profile comment="Test User 8"
/ppp secret add name=user009 password=pass009 service=pppoe profile=pppoe-profile comment="Test User 9"
/ppp secret add name=user010 password=pass010 service=pppoe profile=pppoe-profile comment="Test User 10"

# ============================================================================
# DHCP SERVER CONFIGURATION (for dhcp_server role)
# ============================================================================

# Create IP pool for DHCP
/ip pool add name=dhcp-pool ranges=192.168.100.10-192.168.100.254

# Create DHCP network
/ip dhcp-server network add address=192.168.100.0/24 \
    gateway=192.168.100.1 dns-server=8.8.8.8,8.8.4.4

# Create DHCP server
/ip dhcp-server add name=dhcp-server interface=ether2 \
    address-pool=dhcp-pool disabled=no

# ============================================================================
# NAT CONFIGURATION (for nat_gateway role)
# ============================================================================

# Add masquerade NAT rule
/ip firewall nat add chain=srcnat out-interface=ether1 \
    action=masquerade comment="NAT for outbound traffic"

# ============================================================================
# FIREWALL RULES (for firewall role)
# ============================================================================

# Allow established and related connections
/ip firewall filter add chain=forward connection-state=established,related \
    action=accept comment="Allow established connections"

# Allow ICMP
/ip firewall filter add chain=input protocol=icmp action=accept \
    comment="Allow ICMP"

# Allow management access
/ip firewall filter add chain=input protocol=tcp dst-port=22,80,8728,8729 \
    action=accept comment="Allow management access"

# Allow SNMP
/ip firewall filter add chain=input protocol=udp dst-port=161 \
    action=accept comment="Allow SNMP"

# ============================================================================
# LOGGING CONFIGURATION
# ============================================================================

# Enable logging for PPPoE
/system logging add topics=pppoe,info,debug action=memory

# Enable logging for DHCP
/system logging add topics=dhcp,info action=memory

# Enable logging for firewall
/system logging add topics=firewall,info action=memory

# ============================================================================
# SYSLOG FORWARDING (optional)
# ============================================================================

# Uncomment and configure to forward logs to ISP Monitor
# /system logging action add name=remote remote=172.25.0.1 remote-port=514 target=remote
# /system logging add action=remote topics=system,info
# /system logging add action=remote topics=pppoe,info
# /system logging add action=remote topics=dhcp,info

# ============================================================================
# NETFLOW EXPORT (optional)
# ============================================================================

# Uncomment to enable NetFlow export
# /ip traffic-flow set enabled=yes interfaces=all
# /ip traffic-flow target add address=172.25.0.1:2055 version=9

# ============================================================================
# BANDWIDTH TEST SERVER (for testing)
# ============================================================================

/tool bandwidth-server set enabled=yes authenticate=no

# ============================================================================
# GRAPHING (internal statistics)
# ============================================================================

/tool graphing interface add interface=all store-on-disk=no

# ============================================================================
# RESOURCE MONITORING
# ============================================================================

# Enable system resource monitoring
/tool graphing resource add store-on-disk=no

# ============================================================================
# FINAL STATUS
# ============================================================================

:put "Configuration complete!"
:put "Router Name: [/system identity get name]"
:put "API User: monitor / monitor123"
:put "SNMP Community: public"
:put "API Port: 8728"
:put "API-SSL Port: 8729"
:put ""
:put "Test PPPoE users: user001-user010 / pass001-pass010"
:put ""
:put "You can now add this router to ISP Visual Monitor"

# CHR-CORE-01: Core Router Configuration
# Role: Core routing, BGP peering, main gateway

/system identity
set name=CHR-CORE-01

# ==========================================
# Interface Configuration
# ==========================================
/interface bridge
add name=bridge-mgmt comment="Management Bridge"

# ==========================================
# IP Addressing
# ==========================================
/ip address
add address=172.25.0.10/24 interface=ether1 comment="ISP Network"
add address=172.26.0.10/24 interface=ether2 comment="Monitoring Network"

/ip route
add dst-address=0.0.0.0/0 gateway=172.25.0.1 comment="Default Route"

# ==========================================
# SNMP Configuration (v2c)
# ==========================================
/snmp
set enabled=yes \
    contact="admin@demo-isp.local" \
    location="Core-Datacenter-01" \
    trap-community=public \
    trap-version=2

/snmp community
add name=public \
    addresses=0.0.0.0/0 \
    read-access=yes \
    write-access=no

# ==========================================
# API Configuration
# ==========================================
/ip service
set api address=0.0.0.0/0 disabled=no port=8728
set ssh address=0.0.0.0/0 disabled=no port=22

# ==========================================
# User Management
# ==========================================
/user
add name=monitor password=monitor123 group=read comment="Monitoring User"

# ==========================================
# Syslog Configuration
# ==========================================
/system logging action
add name=remote-syslog \
    remote=172.26.0.1 \
    remote-port=514 \
    src-address=172.26.0.10 \
    target=remote

/system logging
add action=remote-syslog topics=system,info
add action=remote-syslog topics=error,critical

# ==========================================
# Firewall (Allow monitoring)
# ==========================================
/ip firewall filter
add chain=input src-address=172.26.0.0/24 action=accept comment="Allow Monitoring"
add chain=input protocol=icmp action=accept comment="Allow Ping"

# ==========================================
# OSPF Configuration (Optional)
# ==========================================
/routing ospf instance
set [ find default=yes ] router-id=1.1.1.1

/routing ospf area
add name=backbone area-id=0.0.0.0 instance=default

/routing ospf interface-template
add area=backbone interfaces=ether1
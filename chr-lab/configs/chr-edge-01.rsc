# CHR-EDGE-01: Edge Router Configuration
# Role: Edge routing, basic PPPoE, firewall

/system identity
set name=CHR-EDGE-01

# ==========================================
# Interface Configuration
# ==========================================
/interface bridge
add name=bridge-wan comment="WAN Bridge"
add name=bridge-customer comment="Customer Bridge"

# ==========================================
# IP Addressing
# ==========================================
/ip address
add address=172.25.0.11/24 interface=ether1 comment="ISP Core"
add address=172.26.0.11/24 interface=ether2 comment="Monitoring"
add address=172.27.0.11/24 interface=ether3 comment="Customer Network"

# ==========================================
# SNMP Configuration
# ==========================================
/snmp
set enabled=yes \
    contact="admin@demo-isp.local" \
    location="Edge-POP-01" \
    trap-community=public \
    trap-version=2

/snmp community
add name=public addresses=0.0.0.0/0 read-access=yes write-access=no

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
add name=monitor password=monitor123 group=read

# ==========================================
# Firewall (Basic security)
# ==========================================
/ip firewall filter
add chain=input src-address=172.26.0.0/24 action=accept comment="Allow Monitoring"
add chain=input connection-state=established,related action=accept
add chain=input protocol=icmp action=accept
add chain=forward connection-state=established,related action=accept

/ip firewall nat
add chain=srcnat out-interface=ether1 action=masquerade comment="NAT to Core"

# ==========================================
# Syslog Configuration
# ==========================================
/system logging action
add name=remote-syslog remote=172.26.0.1 remote-port=514 src-address=172.26.0.11 target=remote

/system logging
add action=remote-syslog topics=system,info
add action=remote-syslog topics=firewall,info
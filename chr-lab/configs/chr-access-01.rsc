# CHR-ACCESS-01: Access Router Configuration
# Role: NAT, Firewall, Access Control

/system identity
set name=CHR-ACCESS-01

# ==========================================
# IP Addressing
# ==========================================
/ip address
add address=172.25.0.12/24 interface=ether1 comment="ISP Core"
add address=172.26.0.12/24 interface=ether2 comment="Monitoring"

# ==========================================
# SNMP Configuration
# ==========================================
/snmp
set enabled=yes contact="admin@demo-isp.local" location="Access-Router-01"

/snmp community
add name=public addresses=0.0.0.0/0 read-access=yes write-access=no

# ==========================================
# API Configuration
# ==========================================
/ip service
set api address=0.0.0.0/0 disabled=no port=8728

/user
add name=monitor password=monitor123 group=read

# ==========================================
# NAT Configuration
# ==========================================
/ip firewall nat
add chain=srcnat out-interface=ether1 action=masquerade comment="Main NAT"

# ==========================================
# Firewall Rules
# ==========================================
/ip firewall filter
add chain=input src-address=172.26.0.0/24 action=accept comment="Allow Monitoring"
add chain=input connection-state=established,related action=accept
add chain=input protocol=icmp action=accept
add chain=input action=drop comment="Drop all else"

add chain=forward connection-state=established,related action=accept
add chain=forward connection-state=new connection-nat-state=!dstnat action=drop comment="Block non-NAT"
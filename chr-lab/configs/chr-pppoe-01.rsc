# CHR-PPPOE-01: Dedicated PPPoE Server Configuration
# Role: PPPoE authentication, session management

/system identity
set name=CHR-PPPOE-01

# ==========================================
# Interface Configuration
# ==========================================
/interface bridge
add name=bridge-local comment="Local Bridge"

# ==========================================
# IP Addressing
# ==========================================
/ip address
add address=172.27.0.13/24 interface=ether1 comment="Customer Network"
add address=172.26.0.13/24 interface=ether2 comment="Monitoring"

# ==========================================
# PPPoE Server Configuration
# ==========================================

# IP Pool for PPPoE clients
/ip pool
add name=pppoe-pool ranges=10.10.1.2-10.10.255.254 comment="PPPoE IP Pool"

# PPP Profile
/ppp profile
add name=pppoe-profile \
    local-address=10.10.0.1 \
    remote-address=pppoe-pool \
    use-compression=no \
    use-encryption=no \
    only-one=no \
    comment="Default PPPoE Profile"

# Enable PPPoE Server
/interface pppoe-server server
add service-name=ISP-DEMO-PPPOE \
    interface=ether1 \
    default-profile=pppoe-profile \
    authentication=pap,chap,mschap1,mschap2 \
    keepalive-timeout=60 \
    max-sessions=1000 \
    max-mtu=1480 \
    max-mru=1480 \
    disabled=no

# PPPoE Users (Test customers)
/ppp secret
add name=customer001 password=pass001 profile=pppoe-profile comment="Test Customer 1"
add name=customer002 password=pass002 profile=pppoe-profile comment="Test Customer 2"
add name=customer003 password=pass003 profile=pppoe-profile comment="Test Customer 3"
add name=customer004 password=pass004 profile=pppoe-profile comment="Test Customer 4"
add name=customer005 password=pass005 profile=pppoe-profile comment="Test Customer 5"
add name=customer006 password=pass006 profile=pppoe-profile comment="Test Customer 6"
add name=customer007 password=pass007 profile=pppoe-profile comment="Test Customer 7"
add name=customer008 password=pass008 profile=pppoe-profile comment="Test Customer 8"
add name=customer009 password=pass009 profile=pppoe-profile comment="Test Customer 9"
add name=customer010 password=pass010 profile=pppoe-profile comment="Test Customer 10"

# ==========================================
# SNMP Configuration
# ==========================================
/snmp
set enabled=yes \
    contact="admin@demo-isp.local" \
    location="PPPoE-Server-01" \
    trap-community=public \
    trap-version=2 \
    trap-generators=interfaces

/snmp community
add name=public addresses=0.0.0.0/0 read-access=yes write-access=no

# ==========================================
# API Configuration
# ==========================================
/ip service
set api address=0.0.0.0/0 disabled=no port=8728
set api-ssl address=0.0.0.0/0 disabled=no port=8729
set ssh address=0.0.0.0/0 disabled=no port=22

# ==========================================
# User Management
# ==========================================
/user
add name=monitor password=monitor123 group=read comment="Monitoring User"
add name=admin password=admin123 group=full comment="Admin User"

# ==========================================
# Syslog Configuration (PPP events)
# ==========================================
/system logging action
add name=remote-syslog \
    remote=172.26.0.1 \
    remote-port=514 \
    src-address=172.26.0.13 \
    target=remote

/system logging
add action=remote-syslog topics=ppp,info prefix="PPPOE"
add action=remote-syslog topics=ppp,error prefix="PPPOE-ERROR"
add action=remote-syslog topics=system,info
add action=remote-syslog topics=account,info prefix="AUTH"

# ==========================================
# Firewall
# ==========================================
/ip firewall filter
add chain=input src-address=172.26.0.0/24 action=accept comment="Allow Monitoring"
add chain=input connection-state=established,related action=accept
add chain=input protocol=icmp action=accept
add chain=input dst-port=8728 protocol=tcp action=accept comment="Allow API"
add chain=input dst-port=161 protocol=udp action=accept comment="Allow SNMP"

# ==========================================
# Traffic Shaping (Optional)
# ==========================================
/queue simple
add name=customer001-queue target=10.10.1.2/32 max-limit=10M/10M comment="Customer 1 - 10Mbps"
add name=customer002-queue target=10.10.1.3/32 max-limit=20M/20M comment="Customer 2 - 20Mbps"
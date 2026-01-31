# Cloud vs On-Premise Deployment

ISP Visual Monitor supports two primary deployment models to accommodate different customer needs: **Cloud-Hosted (SaaS)** and **On-Premise (Self-Hosted)**.

## Overview

| Feature | Cloud (SaaS) | On-Premise |
|---------|--------------|------------|
| **Hosting** | Managed by us | Customer infrastructure |
| **Updates** | Automatic | Manual (with support) |
| **Data Location** | Our secure cloud | Your data center |
| **License Model** | Monthly subscription | Annual license |
| **Support** | Included | Included (varies by plan) |
| **Scaling** | Automatic | Customer managed |
| **Initial Setup** | Minutes | Hours to days |

## Cloud-Hosted (SaaS)

### Benefits

- **Zero Infrastructure**: No servers to manage
- **Automatic Updates**: Always on the latest version
- **Built-in Scalability**: Handles growth automatically
- **High Availability**: Multi-region deployment
- **Managed Backups**: Daily automated backups

### Configuration

Set the following for cloud deployments:

```bash
DEPLOYMENT_MODE=production
CLOUD_MODE=true
```

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    ISP Monitor Cloud                         │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────┐ │
│  │ Load    │──│ API     │──│ Database│  │ License Server  │ │
│  │ Balancer│  │ Cluster │  │ Cluster │  └─────────────────┘ │
│  └─────────┘  └─────────┘  └─────────┘                       │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ HTTPS
                          ▼
               ┌─────────────────────┐
               │ Customer's Routers  │
               └─────────────────────┘
```

### Data Residency

Cloud deployments store data in our secure data centers. Contact us for:
- EU data residency (GDPR compliance)
- Regional deployment options
- Custom data retention policies

## On-Premise (Self-Hosted)

### Benefits

- **Data Sovereignty**: All data stays in your infrastructure
- **Network Isolation**: Works in air-gapped environments
- **Compliance**: Meet strict regulatory requirements
- **Customization**: Full control over deployment
- **Performance**: Low latency for local router access

### Configuration

Set the following for on-premise deployments:

```bash
DEPLOYMENT_MODE=on-premise
CLOUD_MODE=false
LICENSE_KEY=your-license-key
LICENSE_SERVER_URL=https://license.ispmonitor.com/v1
```

### Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                   Customer Data Center                          │
│  ┌──────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ ISP      │──│ PostgreSQL   │  │ Customer's Routers       │  │
│  │ Monitor  │  │ + PostGIS    │  │ ┌─────┐ ┌─────┐ ┌─────┐  │  │
│  │ API      │──│              │  │ │ R1  │ │ R2  │ │ R3  │  │  │
│  │ + Poller │  └──────────────┘  │ └─────┘ └─────┘ └─────┘  │  │
│  └──────────┘  ┌──────────────┐  └──────────────────────────┘  │
│       │        │ Redis        │                                 │
│       └────────│              │                                 │
│                └──────────────┘                                 │
└────────────────────────────────────────────────────────────────┘
                          │
                          │ License Validation (periodic)
                          ▼
               ┌─────────────────────┐
               │  License Server     │
               │  (Internet)         │
               └─────────────────────┘
```

### License Validation

On-premise deployments require periodic license validation:

1. **Online Validation**: The system contacts our license server to validate your license
2. **Offline Grace Period**: If internet is unavailable, the system continues working for up to 72 hours
3. **Air-Gapped Mode**: Contact us for offline license files for fully disconnected networks

### Hardware Requirements

#### Minimum (Small ISP: 1-50 routers)

- **CPU**: 2 cores
- **RAM**: 4 GB
- **Storage**: 50 GB SSD
- **Network**: 100 Mbps

#### Recommended (Medium ISP: 50-500 routers)

- **CPU**: 4 cores
- **RAM**: 8 GB
- **Storage**: 200 GB SSD
- **Network**: 1 Gbps

#### Enterprise (Large ISP: 500+ routers)

- **CPU**: 8+ cores
- **RAM**: 16+ GB
- **Storage**: 500+ GB SSD (RAID recommended)
- **Network**: 1+ Gbps
- **High Availability**: Multiple nodes recommended

## Feature Comparison by License Plan

### Starter Plan

| Feature | Cloud | On-Premise |
|---------|-------|------------|
| Max Routers | 10 | 10 |
| Max Users | 5 | 5 |
| API Access | ✓ | ✓ |
| Basic Alerts | ✓ | ✓ |
| Standard Support | ✓ | ✓ |

### Professional Plan

| Feature | Cloud | On-Premise |
|---------|-------|------------|
| Max Routers | 100 | 100 |
| Max Users | 25 | 25 |
| API Access | ✓ | ✓ |
| Advanced Alerts | ✓ | ✓ |
| Custom Dashboards | ✓ | ✓ |
| Audit Log | ✓ | ✓ |
| Priority Support | ✓ | ✓ |

### Enterprise Plan

| Feature | Cloud | On-Premise |
|---------|-------|------------|
| Max Routers | Unlimited | Unlimited |
| Max Users | Unlimited | Unlimited |
| API Access | ✓ | ✓ |
| Advanced Alerts | ✓ | ✓ |
| Custom Dashboards | ✓ | ✓ |
| Multi-Tenant | ✓ | ✓ |
| SSO Integration | ✓ | ✓ |
| Audit Log | ✓ | ✓ |
| Priority Support | ✓ | ✓ |
| Dedicated Support | ✓ | ✓ |

## Migration

### Cloud to On-Premise

1. Contact support to request data export
2. Receive encrypted backup file
3. Deploy on-premise infrastructure
4. Import data using provided scripts
5. Update router configurations

### On-Premise to Cloud

1. Create cloud account
2. Request migration assistance
3. Export data using built-in tools
4. Import to cloud instance
5. Verify data integrity
6. Update router configurations

## Security Considerations

### Cloud

- TLS 1.3 encryption in transit
- AES-256 encryption at rest
- SOC 2 Type II compliant
- Regular security audits
- DDoS protection

### On-Premise

- Customer-managed encryption
- Network isolation supported
- VPN/private network compatible
- Customer security policies apply
- Regular security patches provided

## Choosing the Right Model

### Choose Cloud If:

- You want minimal operational overhead
- You need quick deployment
- You don't have dedicated IT infrastructure
- You need automatic scaling
- Data residency isn't a concern

### Choose On-Premise If:

- You have strict data residency requirements
- You need to work in air-gapped environments
- You have existing infrastructure to leverage
- You require deep network integration
- Compliance mandates local data storage

## Support

Both deployment models include support:

- **Documentation**: Comprehensive guides and API reference
- **Email Support**: support@ispmonitor.com
- **Priority Support**: Available for Professional and Enterprise plans
- **On-Site Support**: Available for Enterprise on-premise deployments

---

For questions about choosing the right deployment model, contact sales@ispmonitor.com.

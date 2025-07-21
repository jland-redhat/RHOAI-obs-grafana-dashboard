# Configuration Guide

This document provides detailed information about configuring the RHOAI Observability Grafana Dashboards Helm chart.

## Configuration Parameters

### Basic Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `namespace` | Kubernetes namespace where dashboards will be deployed | `monitoring` | No |
| `grafanaFolder` | Grafana folder name for dashboard organization | `Openshift AI Observability` | No |
| `dashboard_folders` | List of dashboard folders to deploy from `dashboards/` directory | `[llm-d, vllm]` | Yes |

### Grafana Integration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `instanceSelector.matchLabels` | Labels to select target Grafana instance | `{app: grafana}` | Yes |
| `dashboardNamespace` | Dashboard namespace for Grafana organization | `default` | No |
| `plugins` | List of required Grafana plugins | `[]` | No |

### Dashboard Behavior

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `dashboard.refresh` | Default refresh interval | `30s` | No |
| `dashboard.timeFrom` | Default time range | `1h` | No |
| `dashboard.templating.enabled` | Enable template variable substitution | `true` | No |
| `dashboard.tags` | Tags to apply to all dashboards | `[rhoai, observability, ai-ml]` | No |

### RBAC Configuration

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `rbac.create` | Create RBAC resources | `true` | No |
| `rbac.serviceAccountName` | Custom service account name | `""` (auto-generated) | No |

### Grafana Operator

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `grafanaOperator.enabled` | Enable Grafana Operator integration | `true` | No |
| `grafanaOperator.apiVersion` | GrafanaDashboard API version | `grafana.integreatly.org/v1beta1` | No |

### Resource Management

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `resources.limits.cpu` | CPU limit | `100m` | No |
| `resources.limits.memory` | Memory limit | `128Mi` | No |
| `resources.requests.cpu` | CPU request | `50m` | No |
| `resources.requests.memory` | Memory request | `64Mi` | No |

### Labels and Annotations

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `commonLabels` | Labels added to all resources | `{}` | No |
| `commonAnnotations` | Annotations added to all resources | `{}` | No |

## Configuration Examples

### Minimal Configuration

```yaml
# values.yaml
namespace: monitoring
dashboard_folders:
  - llm-d
  - vllm
```

### Production Configuration

```yaml
# values-production.yaml
namespace: rhoai-observability

commonLabels:
  app.kubernetes.io/part-of: rhoai-observability
  environment: production
  team: observability

grafanaFolder: "RHOAI Production Dashboards"

dashboard_folders:
  - llm-d
  - vllm

plugins:
  - name: "grafana-piechart-panel"
    version: "1.6.4"

instanceSelector:
  matchLabels:
    app: grafana
    environment: production

dashboard:
  refresh: "30s"
  timeFrom: "6h"
  templating:
    enabled: true
  tags:
    - "rhoai"
    - "production"
    - "ai-ml"

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Development Configuration

```yaml
# values-dev.yaml
namespace: rhoai-dev

commonLabels:
  environment: development

grafanaFolder: "RHOAI Development"

dashboard_folders:
  - llm-d

instanceSelector:
  matchLabels:
    app: grafana
    environment: development

dashboard:
  refresh: "10s"
  timeFrom: "30m"
  tags:
    - "rhoai"
    - "development"
```

## Plugin Configuration

If your dashboards require specific Grafana plugins, configure them as follows:

```yaml
plugins:
  - name: "grafana-piechart-panel"
    version: "1.6.4"
  - name: "grafana-polystat-panel"
    version: "1.2.10"
  - name: "grafana-clock-panel"
    version: "1.3.0"
```

## Instance Selector Examples

### Single Grafana Instance

```yaml
instanceSelector:
  matchLabels:
    app: grafana
```

### Environment-Specific

```yaml
instanceSelector:
  matchLabels:
    app: grafana
    environment: production
```

### Multiple Label Matching

```yaml
instanceSelector:
  matchLabels:
    app: grafana
    team: observability
    tier: monitoring
```

## Template Variable Substitution

When `dashboard.templating.enabled` is `true`, the chart will automatically replace common template variables:

- `${DS_PROMETHEUS}` → `prometheus`
- `$__rate_interval` → `5m`

You can also configure dashboard-level settings:

```yaml
dashboard:
  refresh: "1m"
  timeFrom: "2h"
  templating:
    enabled: true
```

## Troubleshooting

### Dashboard Not Appearing

1. Check if Grafana Operator is installed:
   ```bash
   kubectl get pods -n grafana-operator-system
   ```

2. Verify GrafanaDashboard resources:
   ```bash
   kubectl get grafanadashboards -n <namespace>
   ```

3. Check instance selector matches your Grafana instance:
   ```bash
   kubectl get grafana -n <grafana-namespace> --show-labels
   ```

### Plugin Issues

If dashboards require specific plugins, ensure they're installed in your Grafana instance or configure them in the values:

```yaml
plugins:
  - name: "required-plugin-name"
    version: "x.y.z"
```

### Permission Issues

If you encounter RBAC issues, ensure the service account has proper permissions:

```yaml
rbac:
  create: true
```

## Best Practices

1. **Environment Separation**: Use different namespaces and instance selectors for different environments
2. **Resource Limits**: Set appropriate resource limits for production deployments
3. **Plugin Management**: Explicitly declare required plugins
4. **Folder Organization**: Use meaningful Grafana folder names
5. **Labeling**: Apply consistent labels for resource management
6. **Testing**: Use the provided test script before production deployment

## Testing Your Configuration

Use the provided test script to validate your configuration:

```bash
./scripts/test-chart.sh
```

This will:
- Lint the chart
- Test template rendering
- Validate YAML syntax
- Check dashboard files
- Package the chart

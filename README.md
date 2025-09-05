# Grafana Dashboards Helm Chart

This Helm chart deploys Grafana dashboards as Kubernetes custom resources using the Grafana Operator.

## Features

- Automatically discovers and deploys all JSON dashboard files from the `dashboards` directory
- Supports templating and variable substitution in dashboard JSON files
- Configurable namespace and labels
- Supports Grafana folder organization
- Customizable refresh intervals and time ranges

## Prerequisites

- Openshift 4.17+
- Helm
- Grafana Operator installed in your cluster

## Installation

1. Add the chart repository:
   ```bash
   helm repo add my-repo https://charts.example.com/
   ```

2. Install the chart:
   ```bash
   helm install my-dashboards my-repo/grafana-dashboards -n monitoring
   ```

## Usage

### Adding Dashboards

1. Place your Grafana dashboard JSON files in the `dashboards` directory
2. The chart will automatically create a `GrafanaDashboard` resource for each JSON file
3. The dashboard name will be derived from the filename (without the .json extension)

### Configuration

The following table lists the configurable parameters of the Grafana Dashboards chart and their default values.

|      Parameter       |                              Description                              |             Default             |
| :------------------: | :-------------------------------------------------------------------: | :-----------------------------: |
|     `namespace`      |            Namespace where the dashboards will be created             |          `monitoring`           |
|    `commonLabels`    |                    Labels to add to all resources                     |              `{}`               |
| `commonAnnotations`  |                  Annotations to add to all resources                  |              `{}`               |
|   `grafanaFolder`    |      Folder name in Grafana where the dashboards will be placed       |            `General`            |
| `dashboard_folders`  |           List of folders inside of `dashboards` to deploy            |              `[]`               |
| `dashboardNamespace` |        Dashboard namespace (used for dashboard identification)        |            `default`            |
|      `plugins`       |          List of Grafana plugins required by the dashboards           |              `[]`               |
|  `instanceSelector`  | Selector for the Grafana instance where dashboards should be deployed | `{matchLabels: {app: grafana}}` |

### Example

```yaml
namespace: monitoring
grafanaFolder: "Kubernetes"
commonLabels:
  app.kubernetes.io/part-of: monitoring
  app.kubernetes.io/managed-by: helm

# List of required Grafana plugins
plugins:
  - name: "grafana-piechart-panel"
    version: "1.6.4"
  - name: "grafana-clock-panel"
    version: "1.3.0"

# Selector for the Grafana instance
instanceSelector:
  matchLabels:
    app: grafana
    # Example with multiple labels:
    # environment: production
    # team: observability

dashboard_folders:
  - llm-d # For LLM-D dashboards
  - vllm # For VLLM dashboards
```

## Creating Dashboards

1. Export your dashboard from Grafana UI or create a new JSON file
2. Save it in the `dashboards` directory with a descriptive name (e.g., `kubernetes-cluster.json`)
3. The chart will automatically pick up the new dashboard on the next deployment

## Creating Datasource

A datasource wtih access to an API key from a service account with access to the Promethus Instance will need to be created.

More info on creating that key can be found [here](https://docs.redhat.com/en/documentation/openshift_container_platform/4.19/html/authentication_and_authorization/understanding-and-creating-service-accounts#service-accounts-overview_understanding-service-accounts)

Creating a new token:
```
oc create token grafana-sa -n grafana-user
```

## Upgrading

To upgrade your deployment with a new dashboard or configuration:

```bash
helm upgrade my-dashboards my-repo/grafana-dashboards -n monitoring -f values.yaml
```

## Uninstalling

To uninstall/delete the deployment:

```bash
helm uninstall my-dashboards -n monitoring
```

## Notes

- The chart uses the `GrafanaDashboard` custom resource which requires the Grafana Operator to be installed in your cluster
- Dashboard JSON files should be valid Grafana dashboard exports
- The chart will automatically convert filenames to kebab-case for resource names

## License

[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0) RHOAI-obs-grafana-dashboard

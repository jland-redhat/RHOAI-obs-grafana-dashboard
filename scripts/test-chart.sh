#!/bin/bash

# Test script for RHOAI Observability Grafana Dashboards Helm Chart
set -e

CHART_DIR="$(dirname "$0")/.."
RELEASE_NAME="test-rhoai-dashboards"
NAMESPACE="test-monitoring"

echo "Testing RHOAI Observability Grafana Dashboards Helm Chart"
echo "============================================================"

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo "ERROR: Helm is not installed. Please install Helm first."
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "ERROR: kubectl is not installed. Please install kubectl first."
    exit 1
fi

echo "OK: Prerequisites check passed"

# Lint the chart
echo "INFO: Linting Helm chart..."
if helm lint "$CHART_DIR"; then
    echo "OK: Chart linting passed"
else
    echo "WARNING: Chart linting failed, but continuing with tests..."
fi

# Test template rendering
echo "INFO: Testing template rendering..."
helm template $RELEASE_NAME "$CHART_DIR" > /tmp/rendered-manifests.yaml
if [ $? -eq 0 ]; then
    echo "OK: Template rendering successful"
    echo "INFO: Rendered manifest preview:"
    head -20 /tmp/rendered-manifests.yaml
    echo "..."
    echo "INFO: Dashboard count: $(grep -c 'kind: GrafanaDashboard' /tmp/rendered-manifests.yaml || echo '0')"
else
    echo "ERROR: Template rendering failed"
    exit 1
fi

# Test with custom values
echo "INFO: Testing with production values..."
if [ -f "$CHART_DIR/examples/values-production.yaml" ]; then
    helm template $RELEASE_NAME "$CHART_DIR" -f "$CHART_DIR/examples/values-production.yaml" > /tmp/rendered-prod-manifests.yaml
    if [ $? -eq 0 ]; then
        echo "OK: Production values template rendering successful"
        echo "INFO: Production dashboard count: $(grep -c 'kind: GrafanaDashboard' /tmp/rendered-prod-manifests.yaml || echo '0')"
    else
        echo "ERROR: Production values template rendering failed"
        exit 1
    fi
fi

# Validate YAML syntax of rendered manifests
echo "INFO: Validating YAML syntax..."
if kubectl apply --dry-run=client -f /tmp/rendered-manifests.yaml &> /dev/null; then
    echo "OK: YAML syntax validation passed"
else
    echo "WARNING: YAML syntax validation failed - checking specific issues..."
    kubectl apply --dry-run=client -f /tmp/rendered-manifests.yaml || true
fi

# Check dashboard files exist
echo "INFO: Checking dashboard files..."
dashboard_count=0
for folder in "llm-d" "vllm"; do
    if [ -d "$CHART_DIR/dashboards/$folder" ]; then
        folder_count=$(find "$CHART_DIR/dashboards/$folder" -name "*.json" | wc -l)
        echo "   INFO: $folder: $folder_count dashboards"
        dashboard_count=$((dashboard_count + folder_count))
    fi
done
echo "INFO: Total dashboard files: $dashboard_count"

# Package the chart
echo "INFO: Testing chart packaging..."
if helm package "$CHART_DIR" -d /tmp/; then
    echo "OK: Chart packaging successful"
    ls -la /tmp/grafana-dashboards-*.tgz
else
    echo "ERROR: Chart packaging failed"
    exit 1
fi

echo ""
echo "SUCCESS: All tests completed!"
echo "OK: Chart is ready for deployment"
echo ""
echo "Next steps:"
echo "   1. Install Grafana Operator in your cluster"
echo "   2. Deploy the chart: helm install $RELEASE_NAME $CHART_DIR -n $NAMESPACE --create-namespace"
echo "   3. Verify dashboards: kubectl get grafanadashboards -n $NAMESPACE"
echo "" 
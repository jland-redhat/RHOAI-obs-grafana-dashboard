package helm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Values represents the Helm chart values structure
type Values struct {
	Namespace          string            `yaml:"namespace"`
	CommonLabels       map[string]string `yaml:"commonLabels"`
	CommonAnnotations  map[string]string `yaml:"commonAnnotations"`
	GrafanaFolder      string            `yaml:"grafanaFolder"`
	DashboardFolders   []string          `yaml:"dashboard_folders"`
	DashboardNamespace string            `yaml:"dashboardNamespace"`
	Plugins            []Plugin          `yaml:"plugins"`
	InstanceSelector   InstanceSelector  `yaml:"instanceSelector"`
	Dashboard          DashboardConfig   `yaml:"dashboard"`
	Resources          Resources         `yaml:"resources"`
	RBAC               RBAC              `yaml:"rbac"`
	GrafanaOperator    GrafanaOperator   `yaml:"grafanaOperator"`
}

// Plugin represents a Grafana plugin
type Plugin struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// InstanceSelector represents the Grafana instance selector
type InstanceSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

// DashboardConfig represents dashboard behavior configuration
type DashboardConfig struct {
	Refresh    string     `yaml:"refresh"`
	TimeFrom   string     `yaml:"timeFrom"`
	Templating Templating `yaml:"templating"`
	Tags       []string   `yaml:"tags"`
}

// Templating represents templating configuration
type Templating struct {
	Enabled bool `yaml:"enabled"`
}

// Resources represents resource limits and requests
type Resources struct {
	Limits   ResourceSpec `yaml:"limits"`
	Requests ResourceSpec `yaml:"requests"`
}

// ResourceSpec represents CPU and memory specifications
type ResourceSpec struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// RBAC represents RBAC configuration
type RBAC struct {
	Create             bool   `yaml:"create"`
	ServiceAccountName string `yaml:"serviceAccountName"`
}

// GrafanaOperator represents Grafana Operator configuration
type GrafanaOperator struct {
	Enabled    bool   `yaml:"enabled"`
	APIVersion string `yaml:"apiVersion"`
}

// LoadValues loads and parses the values.yaml file
func LoadValues(valuesPath string) (*Values, error) {
	data, err := os.ReadFile(valuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %w", err)
	}

	var values Values
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse values file: %w", err)
	}

	return &values, nil
}

// GenerateManifests generates Kubernetes manifests from the Helm chart
func GenerateManifests(chartPath, namespace string) (map[string]interface{}, error) {
	// This would typically use Helm Go SDK or call helm template command
	// For now, we'll use a simplified approach

	valuesPath := filepath.Join(chartPath, "values.yaml")
	values, err := LoadValues(valuesPath)
	if err != nil {
		return nil, err
	}

	manifests := make(map[string]interface{})

	// Generate basic manifest structure
	manifests["apiVersion"] = "v1"
	manifests["kind"] = "List"
	manifests["items"] = []map[string]interface{}{}

	// Add dashboard manifests
	dashboardsPath := filepath.Join(chartPath, "dashboards")
	for _, folder := range values.DashboardFolders {
		folderPath := filepath.Join(dashboardsPath, folder)

		err := filepath.WalkDir(folderPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && filepath.Ext(path) == ".json" {
				dashboardName := filepath.Base(path)
				dashboardName = dashboardName[:len(dashboardName)-5] // remove .json

				manifest := map[string]interface{}{
					"apiVersion": values.GrafanaOperator.APIVersion,
					"kind":       "GrafanaDashboard",
					"metadata": map[string]interface{}{
						"name":      fmt.Sprintf("dashboard-%s", dashboardName),
						"namespace": namespace,
						"labels": map[string]string{
							"app.kubernetes.io/name": "grafana-dashboards",
							"grafana-dashboard":      "true",
							"dashboard-folder":       folder,
						},
					},
					"spec": map[string]interface{}{
						"name":             dashboardName,
						"folder":           values.GrafanaFolder,
						"instanceSelector": values.InstanceSelector,
					},
				}

				items := manifests["items"].([]map[string]interface{})
				manifests["items"] = append(items, manifest)
			}
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to process folder %s: %w", folder, err)
		}
	}

	return manifests, nil
}

// RenderTemplates renders Helm templates using the helm command
func RenderTemplates(chartPath, valuesFile, releaseName, namespace string) (string, error) {
	args := []string{
		"template",
		releaseName,
		chartPath,
		"--namespace", namespace,
	}

	if valuesFile != "" {
		args = append(args, "--values", valuesFile)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("helm template failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// ValidateChart validates the Helm chart structure
func ValidateChart(chartPath string) error {
	// Check required files
	requiredFiles := []string{
		"Chart.yaml",
		"values.yaml",
		"templates",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(chartPath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("required file/directory missing: %s", file)
		}
	}

	// Validate Chart.yaml
	chartFile := filepath.Join(chartPath, "Chart.yaml")
	chartData, err := os.ReadFile(chartFile)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart struct {
		APIVersion string `yaml:"apiVersion"`
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
	}

	if err := yaml.Unmarshal(chartData, &chart); err != nil {
		return fmt.Errorf("invalid Chart.yaml: %w", err)
	}

	if chart.APIVersion == "" || chart.Name == "" || chart.Version == "" {
		return fmt.Errorf("Chart.yaml missing required fields")
	}

	// Validate values.yaml
	_, err = LoadValues(filepath.Join(chartPath, "values.yaml"))
	if err != nil {
		return fmt.Errorf("invalid values.yaml: %w", err)
	}

	return nil
}

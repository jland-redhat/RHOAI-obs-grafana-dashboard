package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rhoai/grafana-dashboards/internal/dashboard"
	"github.com/rhoai/grafana-dashboards/internal/helm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "dashboard-manager",
		Short: "RHOAI Grafana Dashboard Manager",
		Long: `A CLI tool for managing RHOAI Grafana dashboards and Helm chart configurations.
This tool helps validate, generate, and manage Grafana dashboard deployments.`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
	}

	// Add subcommands
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(templateCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func validateCmd() *cobra.Command {
	var chartPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate dashboard JSON files and Helm chart",
		Long:  "Validates all dashboard JSON files for proper structure and Helm chart configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateDashboards(chartPath)
		},
	}

	cmd.Flags().StringVarP(&chartPath, "chart-path", "c", ".", "Path to the Helm chart directory")
	return cmd
}

func generateCmd() *cobra.Command {
	var (
		chartPath    string
		outputFormat string
		namespace    string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Kubernetes manifests from the Helm chart",
		Long:  "Generates Kubernetes manifests that would be created by the Helm chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateManifests(chartPath, outputFormat, namespace)
		},
	}

	cmd.Flags().StringVarP(&chartPath, "chart-path", "c", ".", "Path to the Helm chart directory")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "yaml", "Output format (yaml|json)")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "monitoring", "Target namespace")
	return cmd
}

func listCmd() *cobra.Command {
	var chartPath string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all dashboard files in the chart",
		Long:  "Lists all dashboard JSON files that will be deployed by the Helm chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listDashboards(chartPath)
		},
	}

	cmd.Flags().StringVarP(&chartPath, "chart-path", "c", ".", "Path to the Helm chart directory")
	return cmd
}

func templateCmd() *cobra.Command {
	var (
		chartPath   string
		valuesFile  string
		releaseName string
		namespace   string
	)

	cmd := &cobra.Command{
		Use:   "template",
		Short: "Render Helm templates locally",
		Long:  "Renders Helm templates locally without deploying to cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return renderTemplates(chartPath, valuesFile, releaseName, namespace)
		},
	}

	cmd.Flags().StringVarP(&chartPath, "chart-path", "c", ".", "Path to the Helm chart directory")
	cmd.Flags().StringVarP(&valuesFile, "values", "f", "", "Path to values file")
	cmd.Flags().StringVarP(&releaseName, "release", "r", "test-release", "Release name")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "monitoring", "Target namespace")
	return cmd
}

func validateDashboards(chartPath string) error {
	fmt.Println("INFO: Validating dashboard files...")

	dashboardsPath := filepath.Join(chartPath, "dashboards")

	// Read values.yaml to get dashboard folders
	valuesPath := filepath.Join(chartPath, "values.yaml")
	values, err := helm.LoadValues(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to load values.yaml: %w", err)
	}

	totalDashboards := 0
	var validationErrors []string

	for _, folder := range values.DashboardFolders {
		folderPath := filepath.Join(dashboardsPath, folder)
		fmt.Printf("INFO: Checking folder: %s\n", folder)

		err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && strings.HasSuffix(path, ".json") {
				fmt.Printf("   INFO: Validating: %s\n", filepath.Base(path))

				if err := dashboard.ValidateFile(path); err != nil {
					validationErrors = append(validationErrors, fmt.Sprintf("%s: %v", path, err))
				} else {
					totalDashboards++
				}
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to walk folder %s: %w", folder, err)
		}
	}

	fmt.Printf("\nValidation Summary:\n")
	fmt.Printf("   OK: Valid dashboards: %d\n", totalDashboards)
	fmt.Printf("   ERROR: Validation errors: %d\n", len(validationErrors))

	if len(validationErrors) > 0 {
		fmt.Println("\nValidation Errors:")
		for _, err := range validationErrors {
			fmt.Printf("   - %s\n", err)
		}
		return fmt.Errorf("validation failed with %d errors", len(validationErrors))
	}

	fmt.Println("OK: All dashboards are valid!")
	return nil
}

func generateManifests(chartPath, outputFormat, namespace string) error {
	fmt.Printf("INFO: Generating manifests for namespace: %s\n", namespace)

	manifests, err := helm.GenerateManifests(chartPath, namespace)
	if err != nil {
		return fmt.Errorf("failed to generate manifests: %w", err)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(manifests)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		return encoder.Encode(manifests)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

func listDashboards(chartPath string) error {
	fmt.Println("Dashboard Inventory:")

	dashboardsPath := filepath.Join(chartPath, "dashboards")

	// Read values.yaml to get dashboard folders
	valuesPath := filepath.Join(chartPath, "values.yaml")
	values, err := helm.LoadValues(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to load values.yaml: %w", err)
	}

	totalSize := int64(0)
	totalDashboards := 0

	for _, folder := range values.DashboardFolders {
		folderPath := filepath.Join(dashboardsPath, folder)
		fmt.Printf("\n%s/\n", folder)

		err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && strings.HasSuffix(path, ".json") {
				info, err := d.Info()
				if err != nil {
					return err
				}

				fmt.Printf("   %s (%s)\n",
					filepath.Base(path),
					formatFileSize(info.Size()))

				totalSize += info.Size()
				totalDashboards++
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to list folder %s: %w", folder, err)
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("   Total dashboards: %d\n", totalDashboards)
	fmt.Printf("   Total size: %s\n", formatFileSize(totalSize))

	return nil
}

func renderTemplates(chartPath, valuesFile, releaseName, namespace string) error {
	fmt.Printf("INFO: Rendering templates for release: %s\n", releaseName)

	output, err := helm.RenderTemplates(chartPath, valuesFile, releaseName, namespace)
	if err != nil {
		return fmt.Errorf("failed to render templates: %w", err)
	}

	fmt.Print(output)
	return nil
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Dashboard represents the structure of a Grafana dashboard
type Dashboard struct {
	ID            interface{} `json:"id"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Tags          []string    `json:"tags"`
	Timezone      string      `json:"timezone"`
	Panels        []Panel     `json:"panels"`
	Time          TimeRange   `json:"time"`
	Templating    Templating  `json:"templating"`
	Refresh       interface{} `json:"refresh"`
	SchemaVersion int         `json:"schemaVersion"`
	Version       interface{} `json:"version"`
	UID           string      `json:"uid"`
}

// Panel represents a dashboard panel
type Panel struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Targets     []Target    `json:"targets"`
	GridPos     GridPos     `json:"gridPos"`
	FieldConfig FieldConfig `json:"fieldConfig"`
}

// Target represents a query target
type Target struct {
	Expr           string `json:"expr"`
	RefID          string `json:"refId"`
	IntervalFactor int    `json:"intervalFactor"`
	Step           int    `json:"step"`
}

// GridPos represents panel position
type GridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

// FieldConfig represents field configuration
type FieldConfig struct {
	Defaults FieldDefaults `json:"defaults"`
}

// FieldDefaults represents default field settings
type FieldDefaults struct {
	Unit        string                 `json:"unit"`
	DisplayName string                 `json:"displayName"`
	Custom      map[string]interface{} `json:"custom"`
}

// TimeRange represents the dashboard time range
type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Templating represents dashboard templating
type Templating struct {
	List []TemplateVariable `json:"list"`
}

// TemplateVariable represents a template variable
type TemplateVariable struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Label   string `json:"label"`
	Query   string `json:"query"`
	Refresh int    `json:"refresh"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateFile validates a dashboard JSON file
func ValidateFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var dashboard Dashboard
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return ValidateDashboard(&dashboard)
}

// ValidateDashboard validates a dashboard structure
func ValidateDashboard(dashboard *Dashboard) error {
	var errors []ValidationError

	// Validate required fields
	if dashboard.Title == "" {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "title is required",
		})
	}

	if dashboard.SchemaVersion == 0 {
		errors = append(errors, ValidationError{
			Field:   "schemaVersion",
			Message: "schemaVersion is required",
		})
	}

	// Validate panels
	if len(dashboard.Panels) == 0 {
		errors = append(errors, ValidationError{
			Field:   "panels",
			Message: "at least one panel is required",
		})
	}

	panelIDs := make(map[int]bool)
	for i, panel := range dashboard.Panels {
		// Validate panel ID uniqueness
		if panelIDs[panel.ID] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("panels[%d].id", i),
				Message: fmt.Sprintf("duplicate panel ID: %d", panel.ID),
			})
		}
		panelIDs[panel.ID] = true

		// Validate panel type
		if panel.Type == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("panels[%d].type", i),
				Message: "panel type is required",
			})
		}

		// Validate grid position
		if panel.GridPos.W <= 0 || panel.GridPos.H <= 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("panels[%d].gridPos", i),
				Message: "panel width and height must be positive",
			})
		}

		// Validate targets for query panels
		if isQueryPanel(panel.Type) && len(panel.Targets) == 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("panels[%d].targets", i),
				Message: "query panels must have at least one target",
			})
		}

		// Validate target expressions
		for j, target := range panel.Targets {
			if target.RefID == "" {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("panels[%d].targets[%d].refId", i, j),
					Message: "target refId is required",
				})
			}
		}
	}

	// Validate template variables
	varNames := make(map[string]bool)
	for i, variable := range dashboard.Templating.List {
		if variable.Name == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("templating.list[%d].name", i),
				Message: "variable name is required",
			})
		}

		if varNames[variable.Name] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("templating.list[%d].name", i),
				Message: fmt.Sprintf("duplicate variable name: %s", variable.Name),
			})
		}
		varNames[variable.Name] = true

		if variable.Type == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("templating.list[%d].type", i),
				Message: "variable type is required",
			})
		}
	}

	// Return combined errors
	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
}

// isQueryPanel checks if a panel type requires query targets
func isQueryPanel(panelType string) bool {
	queryPanels := map[string]bool{
		"timeseries": true,
		"stat":       true,
		"gauge":      true,
		"bargauge":   true,
		"table":      true,
		"heatmap":    true,
		"piechart":   true,
		"graph":      true, // legacy
		"singlestat": true, // legacy
	}
	return queryPanels[panelType]
}

// GetDashboardMetadata extracts metadata from a dashboard
func GetDashboardMetadata(dashboard *Dashboard) map[string]interface{} {
	return map[string]interface{}{
		"title":          dashboard.Title,
		"description":    dashboard.Description,
		"uid":            dashboard.UID,
		"tags":           dashboard.Tags,
		"panels_count":   len(dashboard.Panels),
		"schema_version": dashboard.SchemaVersion,
		"has_templating": len(dashboard.Templating.List) > 0,
	}
}

// ProcessTemplateVariables replaces template variables in dashboard content
func ProcessTemplateVariables(content string, replacements map[string]string) string {
	result := content
	for variable, replacement := range replacements {
		// Replace ${VARIABLE} format
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s}", variable), replacement)
		// Replace $VARIABLE format
		result = strings.ReplaceAll(result, fmt.Sprintf("$%s", variable), replacement)
	}
	return result
}

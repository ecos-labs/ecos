package utils

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateRequired checks if a required field is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateURL validates if a string is a valid URL
func ValidateURL(urlStr, fieldName string) error {
	if urlStr == "" {
		return nil // Allow empty URLs if not required
	}

	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%s must be a valid URL: %w", fieldName, err)
	}
	return nil
}

// ValidateFilePath validates if a file path exists and is accessible
func ValidateFilePath(path, fieldName string) error {
	if path == "" {
		return nil // Allow empty paths if not required
	}

	if !filepath.IsAbs(path) {
		// Convert to absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("%s path cannot be resolved: %w", fieldName, err)
		}
		path = absPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s path does not exist: %s", fieldName, path)
	}

	return nil
}

// ValidateDirectoryPath validates if a directory path exists and is accessible
func ValidateDirectoryPath(path, fieldName string) error {
	if path == "" {
		return nil // Allow empty paths if not required
	}

	if err := ValidateFilePath(path, fieldName); err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s path cannot be accessed: %w", fieldName, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory, not a file: %s", fieldName, path)
	}

	return nil
}

// ValidateOneOf validates that a value is one of the allowed options
func ValidateOneOf(value string, options []string, fieldName string) error {
	if value == "" {
		return nil // Allow empty values if not required
	}

	for _, option := range options {
		if value == option {
			return nil
		}
	}

	return fmt.Errorf("%s must be one of: %s (got: %s)", fieldName, strings.Join(options, ", "), value)
}

// ValidateRegex validates that a value matches a regular expression
func ValidateRegex(value, pattern, fieldName string) error {
	if value == "" {
		return nil // Allow empty values if not required
	}

	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("%s regex validation failed: %w", fieldName, err)
	}

	if !matched {
		return fmt.Errorf("%s does not match required pattern", fieldName)
	}

	return nil
}

// ValidatePositiveInt validates that a value is a positive integer
func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive integer (got: %d)", fieldName, value)
	}
	return nil
}

// ValidateNonNegativeInt validates that a value is a non-negative integer
func ValidateNonNegativeInt(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative (got: %d)", fieldName, value)
	}
	return nil
}

// ValidateAWSRegion validates AWS region format
func ValidateAWSRegion(region string) error {
	if region == "" {
		return nil
	}

	// Basic AWS region pattern validation
	pattern := `^[a-z]{2}-[a-z]+-\d{1}$`
	return ValidateRegex(region, pattern, "AWS region")
}

// ValidateGCPProjectID validates GCP project ID format
func ValidateGCPProjectID(projectID string) error {
	if projectID == "" {
		return nil
	}

	// GCP project ID pattern: 6-30 characters, lowercase letters, digits, hyphens
	pattern := `^[a-z][a-z0-9-]{4,28}[a-z0-9]$`
	return ValidateRegex(projectID, pattern, "GCP project ID")
}

// ValidateAzureSubscriptionID validates Azure subscription ID format (UUID)
func ValidateAzureSubscriptionID(subscriptionID string) error {
	if subscriptionID == "" {
		return nil
	}

	// UUID pattern
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	return ValidateRegex(strings.ToLower(subscriptionID), pattern, "Azure subscription ID")
}

// ValidatePluginName validates plugin name format
func ValidatePluginName(name string) error {
	if name == "" {
		return errors.New("plugin name is required")
	}

	// Plugin names should be lowercase with hyphens
	pattern := `^[a-z][a-z0-9-]*[a-z0-9]$`
	return ValidateRegex(name, pattern, "plugin name")
}

// ValidateOutputFormat validates output format options
func ValidateOutputFormat(format string) error {
	validFormats := []string{"json", "yaml", "csv", "table"}
	return ValidateOneOf(format, validFormats, "output format")
}

// ValidateLogLevel validates log level options
func ValidateLogLevel(level string) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	return ValidateOneOf(level, validLevels, "log level")
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("validation failed:\n  - %s", strings.Join(messages, "\n  - "))
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

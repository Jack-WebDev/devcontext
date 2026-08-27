package provider

import (
	"fmt"
	"os"
)

func credentialFileData(path string) ([]byte, bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("inspect provider credential %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, false, fmt.Errorf("provider credential %q is not a regular file", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("read provider credential %q: %w", path, err)
	}
	return data, true, nil
}

func metadataFields(fields ...MetadataField) []MetadataField {
	result := make([]MetadataField, 0, len(fields))
	for _, field := range fields {
		if field.Value != "" {
			result = append(result, field)
		}
	}
	return result
}

func findJSONFieldString(value any, fieldNames ...string) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, name := range fieldNames {
			if value, ok := typed[name].(string); ok && value != "" {
				return value, true
			}
		}
		for _, nested := range typed {
			if value, ok := findJSONFieldString(nested, fieldNames...); ok {
				return value, true
			}
		}
	case []any:
		for _, nested := range typed {
			if value, ok := findJSONFieldString(nested, fieldNames...); ok {
				return value, true
			}
		}
	}
	return "", false
}

func firstJSONFieldString(value any, fieldNames ...string) string {
	found, _ := findJSONFieldString(value, fieldNames...)
	return found
}

func claudeOrganizationName(credentials any) string {
	if name := firstJSONFieldString(credentials, "organizationName", "organizationDisplayName", "organization_name", "organization_display_name", "orgName", "orgDisplayName", "workspaceName", "workspaceDisplayName"); name != "" {
		return name
	}
	return firstStringFieldInNamedJSONObject(credentials, []string{"organization", "organizationInfo", "organization_info", "org", "workspace"}, []string{"name", "displayName", "display_name"})
}

func firstStringFieldInNamedJSONObject(value any, objectNames, fieldNames []string) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, name := range objectNames {
			if nested, ok := typed[name]; ok {
				if found := firstStringField(nested, fieldNames); found != "" {
					return found
				}
			}
		}
		for _, nested := range typed {
			if found := firstStringFieldInNamedJSONObject(nested, objectNames, fieldNames); found != "" {
				return found
			}
		}
	case []any:
		for _, nested := range typed {
			if found := firstStringFieldInNamedJSONObject(nested, objectNames, fieldNames); found != "" {
				return found
			}
		}
	}
	return ""
}

func firstStringField(value any, fieldNames []string) string {
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	for _, name := range fieldNames {
		if value, ok := object[name].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

package apiclient

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Site represents a Sitecore site
type Site struct {
	Name            string `json:"name"`
	HostName        string `json:"hostName,omitempty"`
	TargetHostName  string `json:"targetHostName,omitempty"`
	ContentLanguage struct {
		Name string `json:"name,omitempty"`
	} `json:"contentLanguage,omitempty"`
	Language      string `json:"language,omitempty"`
	Domain        string `json:"domain,omitempty"`
	RootPath      string `json:"rootPath,omitempty"`
	StartPath     string `json:"startPath,omitempty"`
	BrowserTitle  string `json:"browserTitle,omitempty"`
	CacheHtml     bool   `json:"cacheHtml,omitempty"`
	CacheMedia    bool   `json:"cacheMedia,omitempty"`
	EnablePreview bool   `json:"enablePreview,omitempty"`
	RootItem      struct {
		ItemID string `json:"itemId"`
	} `json:"rootItem,omitempty"`
	StartItem struct {
		ItemID string `json:"itemId,omitempty"`
		Path   string `json:"path,omitempty"`
	} `json:"startItem,omitempty"`
}

// SitesResponse represents the response from sites query
type SitesResponse struct {
	Sites []Site `json:"sites"`
}

// Item represents a Sitecore item
type Item struct {
	ItemID       string                 `json:"itemId"`
	Path         string                 `json:"path"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"displayName"`
	TemplateID   string                 `json:"templateId"`
	TemplateName string                 `json:"templateName"`
	Fields       map[string]interface{} `json:"fields"`
	Children     []Item                 `json:"children"`
}

// graphQLItemResponse represents the intermediate structure for GraphQL responses
type graphQLItemResponse struct {
	ItemID      string `json:"itemId"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Template    struct {
		TemplateID string `json:"templateId"`
		Name       string `json:"name"`
	} `json:"template"`
	// DynamicFields captures all field1, field2, etc. properties dynamically
	// This map will contain the raw field data as returned by GraphQL
	DynamicFields map[string]map[string]interface{} `json:"-"`
	Children      struct {
		Nodes []graphQLItemResponse `json:"nodes"`
	} `json:"children"`
}

// UnmarshalJSON implements custom JSON unmarshaling for graphQLItemResponse
// to handle dynamic field1, field2, etc. properties
func (g *graphQLItemResponse) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a temporary struct that can capture all fields
	type Alias graphQLItemResponse
	temp := &struct {
		*Alias
		// Capture dynamic fields - we'll use a large number to handle most cases
		Field1  map[string]interface{} `json:"field1,omitempty"`
		Field2  map[string]interface{} `json:"field2,omitempty"`
		Field3  map[string]interface{} `json:"field3,omitempty"`
		Field4  map[string]interface{} `json:"field4,omitempty"`
		Field5  map[string]interface{} `json:"field5,omitempty"`
		Field6  map[string]interface{} `json:"field6,omitempty"`
		Field7  map[string]interface{} `json:"field7,omitempty"`
		Field8  map[string]interface{} `json:"field8,omitempty"`
		Field9  map[string]interface{} `json:"field9,omitempty"`
		Field10 map[string]interface{} `json:"field10,omitempty"`
		Field11 map[string]interface{} `json:"field11,omitempty"`
		Field12 map[string]interface{} `json:"field12,omitempty"`
		Field13 map[string]interface{} `json:"field13,omitempty"`
		Field14 map[string]interface{} `json:"field14,omitempty"`
		Field15 map[string]interface{} `json:"field15,omitempty"`
		Field16 map[string]interface{} `json:"field16,omitempty"`
		Field17 map[string]interface{} `json:"field17,omitempty"`
		Field18 map[string]interface{} `json:"field18,omitempty"`
		Field19 map[string]interface{} `json:"field19,omitempty"`
		Field20 map[string]interface{} `json:"field20,omitempty"`
	}{
		Alias: (*Alias)(g),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Collect all non-nil dynamic fields into DynamicFields map
	g.DynamicFields = make(map[string]map[string]interface{})

	fields := []map[string]interface{}{
		temp.Field1, temp.Field2, temp.Field3, temp.Field4, temp.Field5,
		temp.Field6, temp.Field7, temp.Field8, temp.Field9, temp.Field10,
		temp.Field11, temp.Field12, temp.Field13, temp.Field14, temp.Field15,
		temp.Field16, temp.Field17, temp.Field18, temp.Field19, temp.Field20,
	}

	for i, field := range fields {
		if len(field) > 0 {
			fieldAlias := fmt.Sprintf("field%d", i+1)
			g.DynamicFields[fieldAlias] = field
		}
	}

	return nil
}

// TransformGraphQLFields transforms GraphQL field structure to uniform string values
// GraphQL returns: fieldName: null OR fieldName: {"value": "actual_value"}
// We want: fieldName: "actual_value" OR fieldName: null (as string)
func TransformGraphQLFields(rawFields map[string]interface{}) map[string]interface{} {
	if rawFields == nil {
		return nil
	}

	result := make(map[string]interface{})

	for fieldName, fieldValue := range rawFields {
		if fieldValue == nil {
			result[fieldName] = nil
		} else if fieldObj, ok := fieldValue.(map[string]interface{}); ok {
			// Handle {"value": "actual_value"} structure
			if value, exists := fieldObj["value"]; exists {
				result[fieldName] = value
			} else {
				result[fieldName] = nil
			}
		} else {
			// Handle direct string values (for backward compatibility)
			result[fieldName] = fieldValue
		}
	}

	return result
}

// FormatItemID converts a Sitecore item ID from the API format (lowercase without braces)
// to the standard GUID format with braces and uppercase letters.
// Example: "87f82eeb362b4923900176b29b448600" -> "{87F82EEB-362B-4923-9001-76B29B448600}"
func FormatItemID(itemID string) string {
	if itemID == "" {
		return itemID
	}

	// Check if the ID is already in the correct format (has braces and is uppercase)
	if strings.HasPrefix(itemID, "{") && strings.HasSuffix(itemID, "}") {
		// Check if it's already properly formatted by checking for dashes
		innerID := strings.Trim(itemID, "{}")
		if strings.Contains(innerID, "-") {
			return itemID // Already in correct format
		}
	}

	// Remove any existing braces
	itemID = strings.Trim(itemID, "{}")

	// Only format if it's exactly 32 characters and contains only hexadecimal characters
	if len(itemID) == 32 {
		// Check if all characters are hexadecimal (0-9, a-f, A-F)
		isHex := true
		for _, c := range itemID {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				isHex = false
				break
			}
		}

		if isHex {
			// Convert to uppercase
			itemID = strings.ToUpper(itemID)

			// Format as GUID: 8-4-4-4-12
			return fmt.Sprintf("{%s-%s-%s-%s-%s}",
				itemID[0:8],
				itemID[8:12],
				itemID[12:16],
				itemID[16:20],
				itemID[20:32])
		}
	}

	// If it doesn't match the expected format, return as-is
	return itemID
}

// ItemResponse represents the response from item queries
type ItemResponse struct {
	Item *Item `json:"item"`
}

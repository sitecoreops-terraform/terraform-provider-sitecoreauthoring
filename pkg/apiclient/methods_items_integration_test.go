package apiclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGraphQLFieldMappingIntegration tests the complete field mapping flow
func TestGetItemByPathIntegration(t *testing.T) {
	t.Run("Integration tests - GetItemByPath", func(t *testing.T) {

		cliEndpointName := "dev"
		client, err := NewClientFromCLIWithEndpoint("", cliEndpointName)
		if err != nil {
			t.Skip("Could not find config to initialize client")
			return
		}
		fieldNames := []string{"Title", "Text"}

		result, err := client.GetItemByPath("/sitecore/content/Home", fieldNames, nil)

		if err != nil {
			t.Skip("Could not execute query, make sure authentication is configured")
			return
		}

		assert.Equal(t, "Home", result.Name)
		assert.Equal(t, "110d559fdea542ea9c1c8a5df7e70ef9", result.ItemID)

		fields := result.Fields
		assert.NotNil(t, fields)
		assert.Len(t, fields, 2)
		assert.Equal(t, "Sitecore Experience Platform", result.Fields["Title"])
	})

	t.Run("Integration tests - GetItemByID", func(t *testing.T) {

		cliEndpointName := "dev"
		client, err := NewClientFromCLIWithEndpoint("", cliEndpointName)
		if err != nil {
			t.Skip("Could not find config to initialize client")
			return
		}
		fieldNames := []string{"Title", "Text"}

		result, err := client.GetItemByID("110d559fdea542ea9c1c8a5df7e70ef9", fieldNames, nil)

		if err != nil {
			t.Skip("Could not execute query, make sure authentication is configured")
			return
		}

		assert.Equal(t, "Home", result.Name)
		assert.Equal(t, "110d559fdea542ea9c1c8a5df7e70ef9", result.ItemID)

		fields := result.Fields
		assert.NotNil(t, fields)
		assert.Len(t, fields, 2)
		assert.Equal(t, "Sitecore Experience Platform", result.Fields["Title"])
	})
}

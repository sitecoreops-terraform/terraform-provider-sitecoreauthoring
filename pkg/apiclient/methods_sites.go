package apiclient

import (
	"encoding/json"
	"fmt"
)

// GetSites fetches all sites from the Sitecore Authoring API
func (c *Client) GetSites() ([]Site, error) {
	query := `
		query {
			sites {
				name
				hostName
				targetHostName
				contentLanguage {
					name
				}
				language
				domain
				rootPath
				startPath
				browserTitle
				cacheMedia
				enablePreview
				rootItem {
					itemId
				}
				startItem {
					itemId
					path
				}
			}
		}
	`

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute sites query: %v", err)
	}

	// Parse the response
	var sitesResponse SitesResponse
	if err := parseGraphQLResponse(response.Data, &sitesResponse); err != nil {
		return nil, fmt.Errorf("failed to parse sites response: %v", err)
	}

	return sitesResponse.Sites, nil
}

// GetSiteByName fetches a specific site by name
func (c *Client) GetSiteByName(name string) (*Site, error) {
	query := fmt.Sprintf(`
		query {
			site(siteName: "%s") {
				name
				hostName
				targetHostName
				contentLanguage {
					name
				}
				language
				domain
				rootPath
				startPath
				browserTitle
				cacheMedia
				enablePreview
				rootItem {
					itemId
				}
				startItem {
					itemId
					path
				}
			}
		}
	`, name)

	response, err := c.doGraphQLRequest(GraphQLRequestOptions{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute site query: %v", err)
	}

	// Parse the response
	var result struct {
		Site *Site `json:"site"`
	}
	if err := parseGraphQLResponse(response.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse site response: %v", err)
	}

	if result.Site == nil {
		return nil, fmt.Errorf("site with name '%s' not found", name)
	}

	return result.Site, nil
}

// parseGraphQLResponse is a helper function to parse GraphQL response data
func parseGraphQLResponse(data map[string]interface{}, result interface{}) error {
	// Convert the map to JSON and then unmarshal to the result
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL data: %v", err)
	}

	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("failed to unmarshal GraphQL data: %v", err)
	}

	return nil
}

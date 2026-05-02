package apiclient

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
